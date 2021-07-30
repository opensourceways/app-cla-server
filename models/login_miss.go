package models

import (
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type LoginMissOption struct {
	LoginMiss *dbmodels.LoginMiss
}

func (lmo *LoginMissOption) DoLoginMiss() IModelError {
	lmo.LoginMiss.MissNum = lmo.LoginMiss.MissNum + 1
	if lmo.LoginMiss.MissNum >= config.AppConfig.AllowLoginMissNum {
		lmo.LoginMiss.LockTime = util.Expiry(config.AppConfig.LockLoginExpiry)
	} else {
		lmo.LoginMiss.LockTime = util.Now()
	}
	return parseDBError(dbmodels.GetDB().UpdateLoginMiss(*lmo.LoginMiss))
}

func (lmo *LoginMissOption) DoLoginSuccess() IModelError {
	lmo.LoginMiss.MissNum = 0
	lmo.LoginMiss.LockTime = util.Now()
	return parseDBError(dbmodels.GetDB().UpdateLoginMiss(*lmo.LoginMiss))
}

func (lmo *LoginMissOption) IsLocked() bool {
	if lmo == nil {
		return false
	}
	return lmo.LoginMiss.IsLocked()
}

func (lmo *LoginMissOption) ResetLockExpiredLoginNum() {
	if lmo.LoginMiss.MissNum >= config.AppConfig.AllowLoginMissNum {
		lmo.LoginMiss.MissNum = 0
	}
}

func InitializeLoginMissOption(linkID, account string) (*LoginMissOption, IModelError) {
	loginMiss, idbError := dbmodels.GetDB().GetLoginMiss(linkID, account)
	if idbError == nil {
		return &LoginMissOption{LoginMiss: loginMiss}, nil
	}

	if idbError.IsErrorOf(dbmodels.ErrNoDBRecord) {
		loginMiss = &dbmodels.LoginMiss{
			LinkID:   linkID,
			Account:  account,
			MissNum:  0,
			LockTime: 0,
		}
		return &LoginMissOption{LoginMiss: loginMiss}, nil
	}

	return nil, newModelError(ErrForGetLoginMiss, idbError)
}
