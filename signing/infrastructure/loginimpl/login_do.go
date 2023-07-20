package loginimpl

import (
	"encoding/json"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

type loginDO struct {
	Frozen    bool `json:"frozen"`
	FailedNum int  `json:"failed_num"`
}

//MarshalBinary in order to store struct directly in redis
func (do *loginDO) MarshalBinary() ([]byte, error) {
	return json.Marshal(do)
}

func (do *loginDO) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, do)
}

func toLoginDo(v *domain.Login) loginDO {
	return loginDO{
		Frozen:    v.Frozen,
		FailedNum: v.FailedNum,
	}
}

func (do *loginDO) toLogin(lid string) domain.Login {
	return domain.Login{
		Id:        lid,
		Frozen:    do.Frozen,
		FailedNum: do.FailedNum,
	}
}
