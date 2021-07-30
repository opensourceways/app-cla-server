package dbmodels

import (
	"github.com/opensourceways/app-cla-server/util"
)

type LoginMiss struct {
	LinkID   string `json:"link_id"`
	Account  string `json:"account"`
	MissNum  int    `json:"miss_num"`
	LockTime int64  `json:"lock_time"`
}

func (lg *LoginMiss) IsLocked() bool {
	if lg == nil {
		return false
	}
	return lg.LockTime >= util.Now()
}
