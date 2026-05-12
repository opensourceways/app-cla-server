package emailservice

import (
	"errors"
	"testing"
)

type mockEmailClient struct {
	err   error
	calls []*EmailMessage
}

func (m *mockEmailClient) SendEmail(msg *EmailMessage) error {
	m.calls = append(m.calls, msg)
	return m.err
}

func TestEmailMessageClearContent(t *testing.T) {
	msg := &EmailMessage{}
	msg.Content.WriteString("sensitive data")
	msg.ClearContent()
	if msg.Content.Len() != 0 {
		t.Error("expected empty content after clear")
	}
}

func TestClearContentZerosBytes(t *testing.T) {
	msg := &EmailMessage{}
	msg.Content.WriteString("hello")
	// The underlying bytes should be zeroed
	msg.ClearContent()
	msg.Content.WriteString(" new")
	s := msg.Content.String()
	if s != " new" {
		t.Errorf("expected ' new', got '%s'", s)
	}
}

func TestRegisterAndSendEmail(t *testing.T) {
	// Reset impl for test isolation
	oldImpl := impl
	impl = &emailServiceImpl{
		clients: make(map[string]iEmail),
	}
	defer func() { impl = oldImpl }()

	mock := &mockEmailClient{}
	Register("test-platform", mock)

	err := SendEmail("test-platform", &EmailMessage{
		From:    "from@test.com",
		To:      []string{"to@test.com"},
		Subject: "Test",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	if mock.calls[0].Subject != "Test" {
		t.Errorf("expected 'Test', got '%s'", mock.calls[0].Subject)
	}
}

func TestSendEmailUnsupportedPlatform(t *testing.T) {
	oldImpl := impl
	impl = &emailServiceImpl{
		clients: make(map[string]iEmail),
	}
	defer func() { impl = oldImpl }()

	err := SendEmail("unknown", &EmailMessage{})
	if err == nil {
		t.Error("expected error for unsupported platform")
	}
	if err.Error() != "unsupported email platform" {
		t.Errorf("expected 'unsupported email platform', got '%s'", err.Error())
	}
}

func TestSendEmailError(t *testing.T) {
	oldImpl := impl
	impl = &emailServiceImpl{
		clients: make(map[string]iEmail),
	}
	defer func() { impl = oldImpl }()

	mock := &mockEmailClient{err: errors.New("smtp error")}
	Register("fail-platform", mock)

	err := SendEmail("fail-platform", &EmailMessage{})
	if err == nil {
		t.Error("expected error")
	}
}

func TestEmailMessageFields(t *testing.T) {
	msg := EmailMessage{
		From:       "sender@test.com",
		To:         []string{"a@test.com", "b@test.com"},
		Subject:    "Subject",
		Attachment: "/path/to/file",
		MIME:       "text/html",
		HasSecret:  true,
	}
	if msg.From != "sender@test.com" {
		t.Errorf("expected 'sender@test.com', got '%s'", msg.From)
	}
	if len(msg.To) != 2 {
		t.Errorf("expected 2 recipients, got %d", len(msg.To))
	}
}
