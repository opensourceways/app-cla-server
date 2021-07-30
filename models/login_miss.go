package models

import (
	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type LoginMissOption struct {
	*dbmodels.LoginMiss
}

func (lmo *LoginMissOption) DoLoginMiss() {
	lmo.MissNum = lmo.MissNum + 1
	if lmo.MissNum >= config.AppConfig.AllowLoginMissNum {
		lmo.LockTime = util.Expiry(config.AppConfig.LockLoginExpiry)
	} else {
		lmo.LockTime = util.Now()
	}
	if err := dbmodels.GetDB().UpdateLoginMiss(*lmo.LoginMiss); err != nil {
		beego.Error(err)
	}

}

func (lmo *LoginMissOption) DoLoginSuccess() {
	lmo.MissNum = 0
	lmo.LockTime = util.Now()
	if err := dbmodels.GetDB().UpdateLoginMiss(*lmo.LoginMiss); err != nil {
		beego.Error(err)
	}
}

func (lmo *LoginMissOption) IsLocked() bool {
	if lmo == nil {
		return false
	}
	return lmo.LoginMiss.IsLocked()
}

func (lmo *LoginMissOption) ResetLockExpiredLoginNum() {
	loginInterval := util.Now() - lmo.LockTime
	if lmo.LoginMiss.MissNum >= config.AppConfig.AllowLoginMissNum ||
		loginInterval > config.AppConfig.LockLoginExpiry {
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
