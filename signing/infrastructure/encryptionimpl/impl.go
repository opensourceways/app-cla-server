package encryptionimpl

import (
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLen       = 16
	iterTimes     = 10000
	encryptKeyLen = 32
)

func NewEncryptionImpl() encryptionImpl {
	return encryptionImpl{}
}

type encryptionImpl struct{}

func (impl encryptionImpl) Encrypt(plainText string) ([]byte, error) {
	salt, err := impl.genSalt()
	if err != nil {
		return nil, err
	}

	return append(salt, impl.encrypt(plainText, salt)...), nil
}

func (impl encryptionImpl) IsSame(plainText string, encrypted []byte) bool {
	if len(encrypted) < saltLen+1 {
		return false
	}

	v := impl.encrypt(plainText, encrypted[:saltLen])
	v1 := encrypted[saltLen:]

	if len(v) != len(v1) {
		return false
	}

	for i := range v {
		if v[i] != v1[i] {
			return false
		}
	}

	return true
}

func (impl encryptionImpl) encrypt(plainText string, salt []byte) []byte {
	return pbkdf2.Key([]byte(plainText), salt, iterTimes, encryptKeyLen, sha256.New)
}

func (impl encryptionImpl) genSalt() ([]byte, error) {
	b := make([]byte, saltLen)
	_, err := rand.Read(b)

	return b, err
}
