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

// send 执行 SMTP 邮件发送（STARTTLS 流程）
//
// 连接流程：
//  1. net.Dial 建立明文 TCP 连接
//  2. smtp.NewClient 创建客户端
//  3. StartTLS 升级为 TLS 加密
//  4. PlainAuth 身份认证
//  5. Mail/Rcpt/Data 发送邮件
//  6. Quit 关闭连接
func (s *SmtpEmailService) send(to, subject, body string) error {
	addr := net.JoinHostPort(s.host, fmt.Sprintf("%d", s.port))

	// 1. 建立 TCP 连接
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server %s: %w", addr, err)
	}

	// 2. 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// 3. STARTTLS 升级为加密连接
	tlsConfig := &tls.Config{
		ServerName:         s.host,
		InsecureSkipVerify: false, // 生产环境必须验证证书
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// 4. 身份认证
	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}

	// 5. 设置发件人
	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// 6. 设置收件人
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// 7. 写入邮件内容
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to create data writer: %w", err)
	}

	msg := buildMessage(s.from, to, subject, body)
	if _, err := w.Write([]byte(msg)); err != nil {
		w.Close()
		return fmt.Errorf("failed to write message body: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return nil
}

// buildMessage 构造 RFC 5322 格式的邮件内容（含中文编码的 Subject）
func buildMessage(from, to, subject, body string) string {
	// 对含中文的 Subject 进行 Base64 编码
	encodedSubject := mime.BEncoding.Encode("utf-8", subject)

	header := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n",
		from, to, encodedSubject,
	)
	return header + body
}

// MockEmailMessage 记录一次邮件调用的信息
type MockEmailMessage struct {
	Type     string // "admin_notification" | "approval_notification"
	Email    string
	Nickname string // 仅 admin_notification
	Approved *bool  // 仅 approval_notification
}

// MockEmailService 模拟邮件服务（用于开发和测试环境，不真正发送邮件）
// 所有调用记录到 Messages 切片中，方便测试断言。
type MockEmailService struct {
	mu       sync.Mutex
	Messages []MockEmailMessage
}

// NewMockEmailService 创建模拟邮件服务实例
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

// SendAdminNotification 模拟发送管理员通知（记录调用到 Messages）
func (m *MockEmailService) SendAdminNotification(newUserEmail, nickname string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, MockEmailMessage{
		Type:     "admin_notification",
		Email:    newUserEmail,
		Nickname: nickname,
	})
	log.Printf("[MockEmail] SendAdminNotification: email=%s, nickname=%s", newUserEmail, nickname)
	return nil
}

// SendApprovalNotification 模拟发送审核结果通知（记录调用到 Messages）
func (m *MockEmailService) SendApprovalNotification(userEmail string, approved bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := approved
	m.Messages = append(m.Messages, MockEmailMessage{
		Type:     "approval_notification",
		Email:    userEmail,
		Approved: &a,
	})
	log.Printf("[MockEmail] SendApprovalNotification: email=%s, approved=%v", userEmail, approved)
	return nil
}
