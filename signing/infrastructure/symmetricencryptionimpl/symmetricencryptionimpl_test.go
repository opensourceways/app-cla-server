package symmetricencryptionimpl

import (
	"testing"
)

func TestNewSymmetricEncryptionImpl(t *testing.T) {
	// AES-128 key must be 16 bytes
	cfg := &Config{EncryptionKey: "1234567890123456"}
	impl, err := NewSymmetricEncryptionImpl(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if impl == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewSymmetricEncryptionImplInvalidKey(t *testing.T) {
	cfg := &Config{EncryptionKey: "short"}
	_, err := NewSymmetricEncryptionImpl(cfg)
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestSymmetricEncryptDecrypt(t *testing.T) {
	cfg := &Config{EncryptionKey: "1234567890123456"}
	impl, err := NewSymmetricEncryptionImpl(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plaintext := []byte("hello world")
	ciphertext, err := impl.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	decrypted, err := impl.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("expected '%s', got '%s'", plaintext, decrypted)
	}
}

func TestDecryptTooShort(t *testing.T) {
	cfg := &Config{EncryptionKey: "1234567890123456"}
	impl, err := NewSymmetricEncryptionImpl(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = impl.Decrypt([]byte("short"))
	if err == nil {
		t.Error("expected error for short ciphertext")
	}
}
