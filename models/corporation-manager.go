package models

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type CorporationManagerAuthentication struct {
	User     string `json:"user"`
	Password string `json:"password"`
	LinkID   string `json:"link_id"`
}

func (this CorporationManagerAuthentication) Validate() IModelError {
	if this.LinkID == "" || this.Password == "" || this.User == "" {
		return newModelError(ErrEmptyPayload, fmt.Errorf("necessary parameters is empty"))
	}

	return nil
}

func (this CorporationManagerAuthentication) Authenticate() (map[string]dbmodels.CorporationManagerCheckResult, IModelError) {
	info := dbmodels.CorporationManagerCheckInfo{Password: this.Password, LinkID: this.LinkID}
	if merr := checkEmailFormat(this.User); merr == nil {
		info.Email = this.User
		info.EmailSuffix = util.EmailSuffix(this.User)
	} else {
		if merr := checkManagerID(this.User); merr != nil {
			return nil, merr
		}

		info.ID = this.User

		i := strings.LastIndex(this.User, "_")
		info.EmailSuffix = this.User[(i + 1):]
	}

	v, err := dbmodels.GetDB().CheckCorporationManagerExist(info)
	if err == nil {
		return v, nil
	}

	return nil, parseDBError(err)
}

func CreateCorporationAdministrator(index SigningIndex, info *dbmodels.CorporationSigningBasicInfo) (
	*dbmodels.CorporationManagerCreateOption, IModelError,
) {
	pw := util.RandStr(8, "alphanum")

	// TODO
	opt := &dbmodels.CorporationManagerCreateOption{
		ID:       fmt.Sprintf("admin_%s", util.EmailSuffix(info.AdminEmail)),
		Name:     info.AdminName,
		Email:    info.AdminEmail,
		Password: pw,
		Role:     dbmodels.RoleAdmin,
	}

	err := dbmodels.GetDB().AddCorpAdministrator(&index, opt)
	if err == nil {
		return opt, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return nil, newModelError(ErrNoLinkOrManagerExists, err)
	}

	return nil, parseDBError(err)
}

type CorporationManagerResetPassword dbmodels.CorporationManagerResetPassword

func (this CorporationManagerResetPassword) Validate() IModelError {
	if this.NewPassword == this.OldPassword {
		return newModelError(ErrSamePassword, fmt.Errorf("the new password is same as old one"))
	}

	return checkPassword(this.NewPassword)
}

func (this CorporationManagerResetPassword) Reset(linkID, email string) IModelError {
	err := dbmodels.GetDB().ResetCorporationManagerPassword(
		linkID, email, dbmodels.CorporationManagerResetPassword(this),
	)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrNoManager, err)
	}
	return parseDBError(err)
}

func GetCorporationDetail(index SigningIndex) (dbmodels.CorporationDetail, IModelError) {
	v, err := dbmodels.GetDB().GetCorporationDetail(&index)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}

	return v, parseDBError(err)
}
