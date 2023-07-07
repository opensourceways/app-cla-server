package domain

import (
	"github.com/opensourceways/app-cla-server/util"
)

type AccessTokenKey struct {
	Id   string
	CSRF string
}

type AccessToken struct {
	Expiry        int64
	Payload       []byte
	EncryptedCSRF []byte
}

func (at *AccessToken) IsValid() bool {
	return at.Expiry >= util.Now()
}
