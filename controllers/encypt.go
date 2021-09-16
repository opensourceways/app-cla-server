package controllers

import (
	"encoding/hex"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/util"
)

func newEncryption() util.SymmetricEncryption {
	e, _ := util.NewSymmetricEncryption(config.AppConfig.SymmetricEncryptionKey, "")
	return e
}

func encryptData(d []byte) (string, error) {
	t, err := newEncryption().Encrypt(d)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(t), nil
}

func decryptData(s string) ([]byte, error) {
	dst, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	return newEncryption().Decrypt(dst)
}
