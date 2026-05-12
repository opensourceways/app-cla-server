package smtpimpl

import (
	"bytes"
	"testing"
)

func TestConfigSetDefault(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefault()
	if cfg.Port != 465 {
		t.Errorf("expected 465, got %d", cfg.Port)
	}
	if cfg.Host != "smtp.exmail.qq.com" {
		t.Errorf("expected 'smtp.exmail.qq.com', got '%s'", cfg.Host)
	}
	if cfg.Platform != "smtp" {
		t.Errorf("expected 'smtp', got '%s'", cfg.Platform)
	}
}

func TestConfigSetDefaultKeepsCustom(t *testing.T) {
	cfg := &Config{
		Port:     587,
		Host:     "smtp.gmail.com",
		Platform: "gmail",
	}
	cfg.SetDefault()
	if cfg.Port != 587 {
		t.Errorf("expected 587, got %d", cfg.Port)
	}
	if cfg.Host != "smtp.gmail.com" {
		t.Errorf("expected 'smtp.gmail.com', got '%s'", cfg.Host)
	}
}

func TestSimpleTxmailMessage(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("This is the email body")

	msg := &EmailMessage{
		From:    "sender@test.com",
		To:      []string{"receiver@test.com"},
		Subject: "Test Subject",
		Content: buf,
		MIME:    "",
	}

	m := simpleTxmailMessage(msg)
	if m == nil {
		t.Fatal("expected non-nil message")
	}
}

func TestSimpleTxmailMessageWithMIME(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("HTML body")

	msg := &EmailMessage{
		From:    "sender@test.com",
		To:      []string{"receiver@test.com"},
		Subject: "HTML Email",
		Content: buf,
		MIME:    "Content-Type: text/html",
	}

	m := simpleTxmailMessage(msg)
	if m == nil {
		t.Fatal("expected non-nil message")
	}
}

func TestCreateTxMailMessageWithoutAttachment(t *testing.T) {
	impl := &smtpImpl{cfg: Config{Port: 587, Host: "smtp.test.com"}}

	var buf bytes.Buffer
	buf.WriteString("body")
	msg := &EmailMessage{
		From:       "from@test.com",
		To:         []string{"to@test.com"},
		Subject:    "Test",
		Content:    buf,
		Attachment: "",
	}

	m, err := impl.createTxMailMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil")
	}
}

func TestCreateTxMailMessageWithAttachment(t *testing.T) {
	impl := &smtpImpl{cfg: Config{Port: 587, Host: "smtp.test.com"}}

	var buf bytes.Buffer
	buf.WriteString("body with attachment")
	msg := &EmailMessage{
		From:       "from@test.com",
		To:         []string{"to@test.com"},
		Subject:    "Test with file",
		Content:    buf,
		Attachment: "/tmp/fake-file.pdf",
	}

	m, err := impl.createTxMailMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil {
		t.Fatal("expected non-nil")
	}
}
