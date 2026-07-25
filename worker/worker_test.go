package worker

import (
	"errors"
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/emailservice"
)

// fakeEmailService 注册到 emailservice，替代真实 SMTP，记录调用次数并可注入返回错误。
type fakeEmailService struct {
	err   error
	calls int
}

func (f *fakeEmailService) SendEmail(msg *emailservice.EmailMessage) error {
	f.calls++
	return f.err
}

// TestTryToSendEmailSuccess 验证 action 首次即成功时返回 true，且 action 仅被调用一次。
func TestTryToSendEmailSuccess(t *testing.T) {
	w := &emailWorker{stop: make(chan struct{})}

	called := 0
	action := func() error {
		called++
		return nil
	}

	if ok := w.tryToSendEmail(action); !ok {
		t.Fatalf("expected true when action succeeds, got false")
	}
	if called != 1 {
		t.Fatalf("expected action to be called once, got %d", called)
	}
}

// TestTryToSendEmailFailureReturnsFalse 验证 action 持续失败时最终返回 false。
// 生产代码每次重试间隔为 1 分钟、最多 10 次，单测若等待耗尽需 ~9 分钟不可接受，
// 故预先关闭 stop 通道使其在首次失败后经 stop 分支快速返回 false（与重试耗尽语义等价）。
func TestTryToSendEmailFailureReturnsFalse(t *testing.T) {
	w := &emailWorker{stop: make(chan struct{})}
	close(w.stop)

	action := func() error {
		return errors.New("send failed")
	}

	if ok := w.tryToSendEmail(action); ok {
		t.Fatalf("expected false when action keeps failing, got true")
	}
}

// TestSendSimpleMessageSuccess 通过注册假 emailservice 使发信成功，
// 覆盖 SendSimpleMessage 中 tryToSendEmail 返回 true 的审计分支（outcome=sent）。
func TestSendSimpleMessageSuccess(t *testing.T) {
	fake := &fakeEmailService{}
	emailservice.Register("test-plat-success", fake)

	w := &emailWorker{stop: make(chan struct{})}
	msg := &EmailMessage{To: []string{"a@example.com"}, Subject: "hi"}

	w.SendSimpleMessage("test-plat-success", msg)
	w.Shutdown() // 关闭 stop 并等待 goroutine 结束（成功路径不依赖 stop）

	if fake.calls != 1 {
		t.Fatalf("expected email service called once, got %d", fake.calls)
	}
}

// TestSendSimpleMessageFailure 通过注册总是失败的假 emailservice，
// 再由 Shutdown 关闭 stop 使重试快速终止，覆盖审计分支（outcome=failed）。
func TestSendSimpleMessageFailure(t *testing.T) {
	fake := &fakeEmailService{err: errors.New("smtp down")}
	emailservice.Register("test-plat-fail", fake)

	w := &emailWorker{stop: make(chan struct{})}
	msg := &EmailMessage{To: []string{"a@example.com"}, Subject: "hi"}

	w.SendSimpleMessage("test-plat-fail", msg)
	w.Shutdown() // 关闭 stop -> 首次失败后经 stop 分支快速返回 false

	if fake.calls < 1 {
		t.Fatalf("expected email service called at least once, got %d", fake.calls)
	}
}
