package randombytesimpl

import (
	"testing"
)

func TestNewRandomBytesImpl(t *testing.T) {
	impl := NewRandomBytesImpl()
	if impl == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNew(t *testing.T) {
	impl := NewRandomBytesImpl()

	b, err := impl.New(16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(b) != 16 {
		t.Errorf("expected 16 bytes, got %d", len(b))
	}

	b2, err := impl.New(32)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(b2) != 32 {
		t.Errorf("expected 32 bytes, got %d", len(b2))
	}

	// Should produce different values
	if len(b) == len(b2) {
		same := true
		for i := range b {
			if b[i] != b2[i] {
				same = false
				break
			}
		}
		if same {
			t.Error("expected different random bytes")
		}
	}
}
