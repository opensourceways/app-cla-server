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
	Platform string `json:"platform"`
	Org      string `json:"org"`
	Repo     string `json:"repo"`
}

func (this CorporationManagerAuthentication) IsValidate() bool {
	if this.Platform == "" || this.Org == "" {
		return false
	}
	return true
}

func (this CorporationManagerAuthentication) OrgIdentity() string {
	po := fmt.Sprintf("%s/%s", this.Platform, this.Org)
	if this.Repo == "" {
		return po
	}
	return fmt.Sprintf("%s/%s", po, this.Repo)
}

func (this CorporationManagerAuthentication) Authenticate() (map[string]dbmodels.CorporationManagerCheckResult, IModelError) {
	info := dbmodels.CorporationManagerCheckInfo{Password: this.Password, OrgIdentity: this.OrgIdentity()}
	if merr := checkEmailFormat(this.User); merr == nil {
		info.Email = this.User
	} else {
		if merr := checkManagerID(this.User); merr != nil {
			return nil, merr
		}

		i := strings.LastIndex(this.User, "_")
		info.EmailSuffix = this.User[(i + 1):]
		info.ID = this.User[:i]
	}

	v, err := dbmodels.GetDB().CheckCorporationManagerExist(info)
	if err == nil {
		return v, nil
	}

	return nil, parseDBError(err)
}

func CreateCorporationAdministrator(linkID, name, email string) (*dbmodels.CorporationManagerCreateOption, IModelError) {
	pw := util.RandStr(8, "alphanum")

	opt := &dbmodels.CorporationManagerCreateOption{
		ID:       "admin",
		Name:     name,
		Email:    email,
		Password: pw,
		Role:     dbmodels.RoleAdmin,
	}
	err := dbmodels.GetDB().AddCorpAdministrator(linkID, opt)
	if err == nil {
		opt.ID = fmt.Sprintf("admin_%s", util.EmailSuffix(email))
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
	return nil
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

func ListCorporationManagers(linkID, email, role string) ([]dbmodels.CorporationManagerListResult, IModelError) {
	v, err := dbmodels.GetDB().ListCorporationManager(linkID, email, role)
	if err == nil {
		if v == nil {
			v = []dbmodels.CorporationManagerListResult{}
		}
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}

	return v, parseDBError(err)
}
