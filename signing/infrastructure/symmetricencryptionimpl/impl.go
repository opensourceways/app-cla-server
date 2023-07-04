package symmetricencryptionimpl

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

func NewSymmetricEncryptionImpl(cfg *Config) (*symmetricEncryptionImpl, error) {
	c, err := aes.NewCipher([]byte(cfg.EncryptionKey))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	return &symmetricEncryptionImpl{aead: gcm}, nil
}

type symmetricEncryptionImpl struct {
	aead cipher.AEAD
}

func (impl *symmetricEncryptionImpl) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, impl.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return impl.aead.Seal(nonce, nonce, plaintext, nil), nil
}

func (impl *symmetricEncryptionImpl) Decrypt(ciphertext []byte) ([]byte, error) {
	nonceSize := impl.aead.NonceSize()
	if len(ciphertext) < nonceSize+1 {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return impl.aead.Open(nil, nonce, ciphertext, nil)
}
