package domain

import (
	"encoding/json"

	"github.com/opensourceways/app-cla-server/util"
)

type AccessTokenKey struct {
	Id   string
	CSRF string
}

type AccessTokenDO struct {
	Expiry        int64  `json:"expiry"`
	Payload       []byte `json:"payload"`
	EncryptedCSRF []byte `json:"encrypted_csrf"`
}

func (at *AccessTokenDO) IsValid() bool {
	return at.Expiry >= util.Now()
}

//MarshalBinary in order to store struct directly in redis
func (at *AccessTokenDO) MarshalBinary() ([]byte, error) {
	return json.Marshal(at)
}

func (at *AccessTokenDO) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, at)
}
