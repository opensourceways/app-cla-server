package dbmodels

import (
	"github.com/opensourceways/app-cla-server/util"
)

type LoginMiss struct {
	LinkID   string `bson:"link_id" json:"link_id" `
	Account  string `bson:"account" json:"account"`
	MissNum  int    `bson:"miss_num" json:"miss_num"`
	LockTime int64  `bson:"lock_time" json:"lock_time"`
}

func (lg *LoginMiss) IsLocked() bool {
	if lg == nil {
		return false
	}
	return lg.LockTime >= util.Now()
}
