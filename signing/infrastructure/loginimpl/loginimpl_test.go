package loginimpl

import (
	"testing"
	"time"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

func TestConfigSetDefault(t *testing.T) {
	cfg := &Config{}
	cfg.SetDefault()
	if cfg.Expire != 300 {
		t.Errorf("expected 300, got %d", cfg.Expire)
	}
}

func TestConfigSetDefaultKeepsCustom(t *testing.T) {
	cfg := &Config{Expire: 600}
	cfg.SetDefault()
	if cfg.Expire != 600 {
		t.Errorf("expected 600, got %d", cfg.Expire)
	}
}

func TestConfigExpire(t *testing.T) {
	cfg := &Config{Expire: 120}
	d := cfg.expire()
	if d != 120*time.Second {
		t.Errorf("expected 120s, got %v", d)
	}
}

func TestToLoginDo(t *testing.T) {
	l := &domain.Login{
		Id:        "test-id",
		Frozen:    true,
		FailedNum: 3,
	}
	do := toLoginDo(l)
	if !do.Frozen {
		t.Error("expected Frozen=true")
	}
	if do.FailedNum != 3 {
		t.Errorf("expected 3, got %d", do.FailedNum)
	}
}

func TestLoginDOToLogin(t *testing.T) {
	do := &loginDO{
		Frozen:    true,
		FailedNum: 5,
	}
	l := do.toLogin("my-id")
	if l.Id != "my-id" {
		t.Errorf("expected 'my-id', got '%s'", l.Id)
	}
	if !l.Frozen {
		t.Error("expected Frozen=true")
	}
	if l.FailedNum != 5 {
		t.Errorf("expected 5, got %d", l.FailedNum)
	}
}

func TestMarshalUnmarshalBinary(t *testing.T) {
	do := &loginDO{
		Frozen:    true,
		FailedNum: 4,
	}
	data, err := do.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var do2 loginDO
	if err := do2.UnmarshalBinary(data); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if do2.Frozen != do.Frozen || do2.FailedNum != do.FailedNum {
		t.Errorf("mismatch: %+v vs %+v", do, do2)
	}
}
