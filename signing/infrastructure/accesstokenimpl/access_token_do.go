package accesstokenimpl

import (
	"encoding/json"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

type AccessTokenDO struct {
	Expiry        int64  `json:"expiry"`
	Payload       []byte `json:"payload"`
	EncryptedCSRF []byte `json:"encrypted_csrf"`
}

//MarshalBinary in order to store struct directly in redis
func (do *AccessTokenDO) MarshalBinary() ([]byte, error) {
	return json.Marshal(do)
}

func (do *AccessTokenDO) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, do)
}

func toAccessTokenDo(v *domain.AccessToken) AccessTokenDO {
	return AccessTokenDO{
		Expiry:        v.Expiry,
		Payload:       v.Payload,
		EncryptedCSRF: v.EncryptedCSRF,
	}
}

func (do *AccessTokenDO) toAccessToken() domain.AccessToken {
	return domain.AccessToken{
		Expiry:        do.Expiry,
		Payload:       do.Payload,
		EncryptedCSRF: do.EncryptedCSRF,
	}
}
