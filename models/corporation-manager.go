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

func CreateCorporationAdministratorByAdapter(csId string) (dbmodels.CorporationManagerCreateOption, IModelError) {
	return corpAdminAdatperInstance.Add(csId)
}

func CreateCorporationAdministrator(linkID, name, email string) (dbmodels.CorporationManagerCreateOption, IModelError) {
	pw := util.RandStr(8, "alphanum")

	opt := dbmodels.CorporationManagerCreateOption{
		ID:       "admin",
		Name:     name,
		Email:    email,
		Password: pw,
		Role:     dbmodels.RoleAdmin,
	}
	err := dbmodels.GetDB().AddCorpAdministrator(linkID, &opt)
	if err == nil {
		opt.ID = fmt.Sprintf("admin_%s", util.EmailSuffix(email))
		return opt, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return opt, newModelError(ErrNoLinkOrManagerExists, err)
	}

	return opt, parseDBError(err)
}

type CorporationManagerChangePassword dbmodels.CorporationManagerChangePassword

func (opt *CorporationManagerChangePassword) ChangePassword(index string) IModelError {
	return userAdapterInstance.ChangePassword(index, opt)
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

type CorpManagerLoginInfo struct {
	Role             string
	Email            string
	UserId           string
	CorpName         string
	SigningId        string
	InitialPWChanged bool
}

func CorpManagerLogin(opt *CorporationManagerAuthentication) (CorpManagerLoginInfo, IModelError) {
	return userAdapterInstance.Login(opt)
}
