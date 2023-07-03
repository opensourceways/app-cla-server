package encryptionimpl

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLen       = 16
	encryptKeyLen = 32
	iter          = 10000
)

func NewEncryptionImpl() *encryptionImpl {
	return &encryptionImpl{}
}

type encryptionImpl struct{}

func (impl *encryptionImpl) Encrypt(pwd string) (string, error) {
	salt, err := impl.genSalt()
	if err != nil {
		return "", err
	}

	return salt + impl.encrypt(pwd, salt), nil
}

func (impl *encryptionImpl) Check(pwd, encryptStr string) bool {
	enPwd := encryptStr[saltLen:]
	salt := encryptStr[0:saltLen]

	return enPwd == impl.encrypt(pwd, salt)
}

func (impl *encryptionImpl) encrypt(pwd, salt string) string {
	dk := pbkdf2.Key([]byte(pwd), []byte(salt), iter, encryptKeyLen, sha256.New)
	return hex.EncodeToString(dk)
}

func (impl *encryptionImpl) genSalt() (string, error) {
	b := make([]byte, saltLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	encodeSalt := base64.StdEncoding.EncodeToString(b)

	return encodeSalt[0:saltLen], nil
}
