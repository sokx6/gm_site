package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"mime"
	"net"
	"net/smtp"
	"sync"
)

// EmailService 邮件服务接口
type EmailService interface {
	// SendAdminNotification 发送新用户注册通知给管理员
	SendAdminNotification(newUserEmail, nickname string) error
	// SendApprovalNotification 发送审核结果通知给用户
	SendApprovalNotification(userEmail string, approved bool) error
	// SendFriendRequestNotification 发送好友请求通知
	SendFriendRequestNotification(toEmail, fromNickname string) error
	// SendFriendAcceptedNotification 发送好友请求被接受通知
	SendFriendAcceptedNotification(toEmail, fromNickname string) error
	// SendFriendRejectedNotification 发送好友请求被拒绝通知
	SendFriendRejectedNotification(toEmail, fromNickname string) error
	// SendRegisterApprovedNotification 发送注册审核通过通知
	SendRegisterApprovedNotification(toEmail string) error
	// SendRegisterRejectedNotification 发送注册审核拒绝通知
	SendRegisterRejectedNotification(toEmail string) error
	// SendCommentReplyNotification 发送评论回复通知
	SendCommentReplyNotification(toEmail, replyNickname, imageTitle string) error
	// SendImageCommentNotification 发送图片评论通知
	SendImageCommentNotification(toEmail, commenterName, imageTitle, commentContent string) error
}

// SmtpEmailService SMTP 邮件服务实现（587 端口 STARTTLS）
type SmtpEmailService struct {
	host       string
	port       int
	username   string
	password   string
	from       string
	adminEmail string
}

// NewSmtpEmailService 创建 SMTP 邮件服务实例
func NewSmtpEmailService(host string, port int, username, password, from, adminEmail string) *SmtpEmailService {
	return &SmtpEmailService{
		host:       host,
		port:       port,
		username:   username,
		password:   password,
		from:       from,
		adminEmail: adminEmail,
	}
}

// SendAdminNotification 发送新用户注册通知给管理员
func (s *SmtpEmailService) SendAdminNotification(newUserEmail, nickname string) error {
	subject := "新用户注册待审核"
	body := fmt.Sprintf("新用户注册信息：\n\n邮箱：%s\n昵称：%s\n\n请登录管理后台审核。", newUserEmail, nickname)
	return s.send(s.adminEmail, subject, body)
}

// SendApprovalNotification 发送审核结果通知给用户
func (s *SmtpEmailService) SendApprovalNotification(userEmail string, approved bool) error {
	subject := "账号审核结果"
	var body string
	if approved {
		body = "您好！\n\n您的账号已通过审核，现在可以正常登录并使用所有功能。\n\n感谢您的支持！"
	} else {
		body = "您好！\n\n很抱歉，您的账号未通过审核。如有疑问，请联系管理员。\n\n感谢您的理解！"
	}
	return s.send(userEmail, subject, body)
}

// SendRegisterApprovedNotification 发送注册审核通过通知
func (s *SmtpEmailService) SendRegisterApprovedNotification(userEmail string) error {
	subject := "账号审核通过"
	body := "你的账号已通过审核，现在可以登录使用全部功能。"
	return s.send(userEmail, subject, body)
}

// SendRegisterRejectedNotification 发送注册审核拒绝通知
func (s *SmtpEmailService) SendRegisterRejectedNotification(userEmail string) error {
	subject := "账号审核未通过"
	body := "你的账号未通过审核，如有疑问请联系管理员。"
	return s.send(userEmail, subject, body)
}

// SendFriendRequestNotification 发送好友请求通知
func (s *SmtpEmailService) SendFriendRequestNotification(toEmail, fromNickname string) error {
	subject := "好友请求"
	body := fmt.Sprintf("%s 向你发送了好友请求，请登录查看。", fromNickname)
	return s.send(toEmail, subject, body)
}

// SendFriendAcceptedNotification 发送好友请求已通过通知
func (s *SmtpEmailService) SendFriendAcceptedNotification(toEmail, fromNickname string) error {
	subject := "好友请求已通过"
	body := fmt.Sprintf("%s 已接受你的好友请求。", fromNickname)
	return s.send(toEmail, subject, body)
}

// SendFriendRejectedNotification 发送好友请求已拒绝通知
func (s *SmtpEmailService) SendFriendRejectedNotification(toEmail, fromNickname string) error {
	subject := "好友请求已拒绝"
	body := fmt.Sprintf("%s 已拒绝你的好友请求。", fromNickname)
	return s.send(toEmail, subject, body)
}

// SendCommentReplyNotification 发送评论回复通知
func (s *SmtpEmailService) SendCommentReplyNotification(toEmail, replyNickname, imageTitle string) error {
	subject := "评论被回复"
	body := fmt.Sprintf("%s 回复了你在《%s》中的评论。", replyNickname, imageTitle)
	return s.send(toEmail, subject, body)
}

// SendImageCommentNotification 发送图片评论通知
func (s *SmtpEmailService) SendImageCommentNotification(toEmail, commenterName, imageTitle, commentContent string) error {
	subject := "你的图片有新评论"
	body := fmt.Sprintf("%s 评论了你上传的《%s》：%s", commenterName, imageTitle, commentContent)
	return s.send(toEmail, subject, body)
}

// send 执行 SMTP 邮件发送
//
// 端口 465: 直接 TLS 连接（tls.Dial），然后 SMTP 明文认证
// 其他端口: TCP 明文连接 + STARTTLS 升级
func (s *SmtpEmailService) send(to, subject, body string) error {
	addr := net.JoinHostPort(s.host, fmt.Sprintf("%d", s.port))

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	msg := buildMessage(s.from, to, subject, body)

	var client *smtp.Client
	var conn net.Conn
	var err error

	if s.port == 465 {
		tlsConfig := &tls.Config{
			ServerName:         s.host,
			InsecureSkipVerify: false,
		}
		conn, err = tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect TLS to SMTP server %s: %w", addr, err)
		}
		defer conn.Close()

		client, err = smtp.NewClient(conn, s.host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	} else {
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server %s: %w", addr, err)
		}

		client, err = smtp.NewClient(conn, s.host)
		if err != nil {
			conn.Close()
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Quit()

		tlsConfig := &tls.Config{
			ServerName:         s.host,
			InsecureSkipVerify: false,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	if err = client.Mail(s.from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to create data writer: %w", err)
	}

	if _, err = w.Write([]byte(msg)); err != nil {
		w.Close()
		return fmt.Errorf("failed to write message body: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	if s.port != 465 {
		return client.Quit()
	}
	return client.Close()
}

// buildMessage 构造 RFC 5322 格式的邮件内容（含中文编码的 Subject）
func buildMessage(from, to, subject, body string) string {
	encodedSubject := mime.BEncoding.Encode("utf-8", subject)

	header := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n",
		from, to, encodedSubject,
	)
	return header + body
}

// MockEmailMessage 记录一次邮件调用的信息
type MockEmailMessage struct {
	Type         string
	Email        string
	Nickname     string
	Approved     *bool
	FromNickname string
	ImageTitle   string
}

// MockEmailService 模拟邮件服务（用于开发和测试环境，不真正发送邮件）
type MockEmailService struct {
	mu       sync.Mutex
	Messages []MockEmailMessage
}

// NewMockEmailService 创建模拟邮件服务实例
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

// SendAdminNotification 模拟发送管理员通知
func (m *MockEmailService) SendAdminNotification(newUserEmail, nickname string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, MockEmailMessage{
		Type:     "admin_notification",
		Email:    newUserEmail,
		Nickname: nickname,
	})
	log.Printf("[MockEmail] admin notification for %s (nickname: %s)", newUserEmail, nickname)
	return nil
}

// SendApprovalNotification 模拟发送审核结果通知
func (m *MockEmailService) SendApprovalNotification(userEmail string, approved bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := approved
	m.Messages = append(m.Messages, MockEmailMessage{
		Type:     "approval_notification",
		Email:    userEmail,
		Approved: &a,
	})
	log.Printf("[MockEmail] approval notification for %s (approved: %v)", userEmail, approved)
	return nil
}

// SendRegisterApprovedNotification 模拟发送注册审核通过通知
func (m *MockEmailService) SendRegisterApprovedNotification(userEmail string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, MockEmailMessage{
		Type:  "register_approved",
		Email: userEmail,
	})
	log.Printf("[MockEmail] register approved notification for %s", userEmail)
	return nil
}

// SendRegisterRejectedNotification 模拟发送注册审核拒绝通知
func (m *MockEmailService) SendRegisterRejectedNotification(userEmail string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, MockEmailMessage{
		Type:  "register_rejected",
		Email: userEmail,
	})
	log.Printf("[MockEmail] register rejected notification for %s", userEmail)
	return nil
}

// SendFriendRequestNotification 模拟发送好友请求通知
func (m *MockEmailService) SendFriendRequestNotification(toEmail, fromNickname string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, MockEmailMessage{
		Type:         "friend_request",
		Email:        toEmail,
		FromNickname: fromNickname,
	})
	log.Printf("[MockEmail] friend request notification for %s from %s", toEmail, fromNickname)
	return nil
}

// SendFriendAcceptedNotification 模拟发送好友请求已通过通知
func (m *MockEmailService) SendFriendAcceptedNotification(toEmail, fromNickname string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, MockEmailMessage{
		Type:         "friend_accepted",
		Email:        toEmail,
		FromNickname: fromNickname,
	})
	log.Printf("[MockEmail] friend accepted notification for %s from %s", toEmail, fromNickname)
	return nil
}

// SendFriendRejectedNotification 模拟发送好友请求已拒绝通知
func (m *MockEmailService) SendFriendRejectedNotification(toEmail, fromNickname string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, MockEmailMessage{
		Type:         "friend_rejected",
		Email:        toEmail,
		FromNickname: fromNickname,
	})
	log.Printf("[MockEmail] friend rejected notification for %s from %s", toEmail, fromNickname)
	return nil
}

// SendCommentReplyNotification 模拟发送评论回复通知
func (m *MockEmailService) SendCommentReplyNotification(toEmail, replyNickname, imageTitle string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, MockEmailMessage{
		Type:         "comment_reply",
		Email:        toEmail,
		FromNickname: replyNickname,
		ImageTitle:   imageTitle,
	})
	log.Printf("[MockEmail] comment reply notification for %s (replier: %s, image: %s)", toEmail, replyNickname, imageTitle)
	return nil
}

// SendImageCommentNotification 模拟发送图片评论通知
func (m *MockEmailService) SendImageCommentNotification(toEmail, commenterName, imageTitle, commentContent string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, MockEmailMessage{
		Type:         "image_comment",
		Email:        toEmail,
		FromNickname: commenterName,
		ImageTitle:   imageTitle,
	})
	log.Printf("[MockEmail] image comment notification for %s (commenter: %s, image: %s)", toEmail, commenterName, imageTitle)
	return nil
}
