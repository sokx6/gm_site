package service

import (
	"log/slog"
	"os"
	"strconv"
	"testing"

	"gm_site/internal/logger"
)

// TestMain initializes the global logger for service tests.
func TestMain(m *testing.M) {
	logger.L = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	os.Exit(m.Run())
}

// compile-time interface implementation check
var _ EmailService = (*MockEmailService)(nil)
var _ EmailService = (*SmtpEmailService)(nil)

// TestMockEmailService 验证 MockEmailService 正常工作
func TestMockEmailService(t *testing.T) {
	svc := NewMockEmailService()

	// SendAdminNotification 不应返回错误
	if err := svc.SendAdminNotification("user@test.com", "testuser"); err != nil {
		t.Errorf("SendAdminNotification should not error, got: %v", err)
	}

	// SendApprovalNotification（审核通过）不应返回错误
	if err := svc.SendApprovalNotification("user@test.com", true); err != nil {
		t.Errorf("SendApprovalNotification(approved=true) should not error, got: %v", err)
	}

	// SendApprovalNotification（审核拒绝）不应返回错误
	if err := svc.SendApprovalNotification("user@test.com", false); err != nil {
		t.Errorf("SendApprovalNotification(approved=false) should not error, got: %v", err)
	}
}

// TestSmtpEmailService_Construct 验证 SmtpEmailService 构造不 panic
func TestSmtpEmailService_Construct(t *testing.T) {
	svc := NewSmtpEmailService("smtp.example.com", 587, "user", "pass", "from@test.com", "admin@test.com")
	if svc == nil {
		t.Fatal("NewSmtpEmailService should not return nil")
	}
	if svc.host != "smtp.example.com" {
		t.Errorf("expected host smtp.example.com, got %s", svc.host)
	}
	if svc.port != 587 {
		t.Errorf("expected port 587, got %d", svc.port)
	}
}

// TestSmtpEmailService_Send 真实 SMTP 发送测试（需要配置环境变量）
//
// 运行方式：
//
//	GM_SMTP_HOST=smtp.163.com GM_SMTP_PORT=587 \
//	GM_SMTP_USERNAME=xxx@163.com GM_SMTP_PASSWORD=xxx \
//	GM_SMTP_FROM=xxx@163.com GM_SMTP_ADMIN_EMAIL=admin@example.com \
//	go test -v -run TestSmtpEmailService_Send
func TestSmtpEmailService_Send(t *testing.T) {
	host := os.Getenv("GM_SMTP_HOST")
	portStr := os.Getenv("GM_SMTP_PORT")
	username := os.Getenv("GM_SMTP_USERNAME")
	password := os.Getenv("GM_SMTP_PASSWORD")
	from := os.Getenv("GM_SMTP_FROM")
	adminEmail := os.Getenv("GM_SMTP_ADMIN_EMAIL")

	if host == "" || username == "" || password == "" {
		t.Skip("SKIP: SMTP credentials not configured via environment variables")
	}

	port := 587
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	svc := NewSmtpEmailService(host, port, username, password, from, adminEmail)

	// 测试发送管理员通知
	t.Run("SendAdminNotification", func(t *testing.T) {
		if err := svc.SendAdminNotification("newuser@test.com", "测试用户"); err != nil {
			t.Fatalf("SendAdminNotification failed: %v", err)
		}
	})

	// 测试发送审核通过通知
	t.Run("SendApprovalNotification_Approved", func(t *testing.T) {
		if err := svc.SendApprovalNotification("testuser@test.com", true); err != nil {
			t.Fatalf("SendApprovalNotification(approved=true) failed: %v", err)
		}
	})

	// 测试发送审核拒绝通知
	t.Run("SendApprovalNotification_Rejected", func(t *testing.T) {
		if err := svc.SendApprovalNotification("testuser@test.com", false); err != nil {
			t.Fatalf("SendApprovalNotification(approved=false) failed: %v", err)
		}
	})
}


