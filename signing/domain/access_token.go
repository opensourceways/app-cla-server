package domain

import "github.com/opensourceways/app-cla-server/util"

func NewAccessToken(payload, csrf []byte) AccessToken {
	return AccessToken{
		Expiry:        config.AccessTokenExpiry + util.Now(),
		Payload:       payload,
		EncryptedCSRF: csrf,
	}
}

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
