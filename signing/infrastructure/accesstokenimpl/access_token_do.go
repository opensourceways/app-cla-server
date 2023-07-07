package accesstokenimpl

import (
	"encoding/json"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

type accessTokenDO struct {
	Expiry        int64  `json:"expiry"`
	Payload       []byte `json:"payload"`
	EncryptedCSRF []byte `json:"encrypted_csrf"`
}

//MarshalBinary in order to store struct directly in redis
func (do *accessTokenDO) MarshalBinary() ([]byte, error) {
	return json.Marshal(do)
}

func (do *accessTokenDO) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, do)
}

func toAccessTokenDo(v *domain.AccessToken) accessTokenDO {
	return accessTokenDO{
		Expiry:        v.Expiry,
		Payload:       v.Payload,
		EncryptedCSRF: v.EncryptedCSRF,
	}
}

func (do *accessTokenDO) toAccessToken() domain.AccessToken {
	return domain.AccessToken{
		Expiry:        do.Expiry,
		Payload:       do.Payload,
		EncryptedCSRF: do.EncryptedCSRF,
	}
}
