package localclaimpl

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

func TestLocalPath(t *testing.T) {
	impl := &localCLAImpl{dir: "/tmp/cla"}
	p := impl.localPath("link1", "cla1")
	if p != "/tmp/cla/link1_cla1.pdf" {
		t.Errorf("expected '/tmp/cla/link1_cla1.pdf', got '%s'", p)
	}
}

func TestLocalPathOfDiff(t *testing.T) {
	impl := &localCLAImpl{dir: "/tmp/cla"}
	index := &domain.CLAIndex{LinkId: "L1", CLAId: "C2"}
	p := impl.LocalPathOfDiff(index, "signed1")
	if p != "/tmp/cla/L1_signed1_C2.txt" {
		t.Errorf("expected '/tmp/cla/L1_signed1_C2.txt', got '%s'", p)
	}
}

func TestLocalPathOfDiff2(t *testing.T) {
	// Verify the method is consistent
	impl := &localCLAImpl{dir: "/data/cla"}
	index := &domain.CLAIndex{LinkId: "link-123", CLAId: "cla-456"}
	p := impl.LocalPathOfDiff(index, "old789")
	if p != "/data/cla/link-123_old789_cla-456.txt" {
		t.Errorf("expected '/data/cla/link-123_old789_cla-456.txt', got '%s'", p)
	}
}

func TestLocalPath2(t *testing.T) {
	impl := &localCLAImpl{dir: "/var/lib/cla"}
	p := impl.localPath("abc", "001")
	if p != "/var/lib/cla/abc_001.pdf" {
		t.Errorf("expected '/var/lib/cla/abc_001.pdf', got '%s'", p)
	}
}

func TestNewLocalCLAImpl(t *testing.T) {
	impl := NewLocalCLAImpl(&Config{Dir: "/tmp"})
	if impl == nil {
		t.Fatal("expected non-nil")
	}
	if impl.dir != "/tmp" {
		t.Errorf("expected '/tmp', got '%s'", impl.dir)
	}
}
