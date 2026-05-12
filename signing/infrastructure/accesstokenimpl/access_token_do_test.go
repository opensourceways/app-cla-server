package accesstokenimpl

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

func TestToAccessTokenDo(t *testing.T) {
	at := &domain.AccessToken{
		Expiry:        100,
		Payload:       []byte("payload"),
		EncryptedCSRF: []byte("csrf"),
	}
	do := toAccessTokenDo(at)
	if do.Expiry != 100 {
		t.Errorf("expected 100, got %d", do.Expiry)
	}
	if string(do.Payload) != "payload" {
		t.Errorf("expected 'payload', got '%s'", string(do.Payload))
	}
	if string(do.EncryptedCSRF) != "csrf" {
		t.Errorf("expected 'csrf', got '%s'", string(do.EncryptedCSRF))
	}
}

func TestAccessTokenDOToAccessToken(t *testing.T) {
	do := &accessTokenDO{
		Expiry:        200,
		Payload:       []byte("p"),
		EncryptedCSRF: []byte("c"),
	}
	at := do.toAccessToken()
	if at.Expiry != 200 {
		t.Errorf("expected 200, got %d", at.Expiry)
	}
	if string(at.Payload) != "p" {
		t.Errorf("expected 'p', got '%s'", string(at.Payload))
	}
}

func TestMarshalUnmarshalBinary(t *testing.T) {
	do := &accessTokenDO{
		Expiry:        300,
		Payload:       []byte("test-payload"),
		EncryptedCSRF: []byte("test-csrf"),
	}
	data, err := do.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var do2 accessTokenDO
	if err := do2.UnmarshalBinary(data); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if do2.Expiry != do.Expiry {
		t.Errorf("expected %d, got %d", do.Expiry, do2.Expiry)
	}
	if string(do2.Payload) != string(do.Payload) {
		t.Errorf("payload mismatch: '%s' vs '%s'", do2.Payload, do.Payload)
	}
}
