package encryptionimpl

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	impl := NewEncryptionImpl()

	plaintext := []byte("test password")
	encrypted, err := impl.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	if len(encrypted) <= saltLen {
		t.Error("encrypted data should include salt")
	}

	if !impl.IsSame(plaintext, encrypted) {
		t.Error("expected IsSame to return true")
	}

	if impl.IsSame([]byte("wrong"), encrypted) {
		t.Error("expected IsSame to return false for wrong plaintext")
	}
}

func TestIsSameShortCiphertext(t *testing.T) {
	impl := NewEncryptionImpl()

	// ciphertext shorter than saltLen+1
	if impl.IsSame([]byte("test"), []byte("short")) {
		t.Error("expected false for short ciphertext")
	}
}

func TestNewEncryptionImpl(t *testing.T) {
	impl := NewEncryptionImpl()
	// encryptionImpl is a struct value, verify it can encrypt/decrypt
	_, err := impl.Encrypt([]byte("test"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
