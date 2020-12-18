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
}

func (this CorporationManagerAuthentication) Authenticate() (map[string]dbmodels.CorporationManagerCheckResult, error) {
	info := dbmodels.CorporationManagerCheckInfo{Password: this.Password}
	if merr := checkEmailFormat(this.User); merr == nil {
		info.Email = this.User
	} else {
		if _, err := checkManagerID(this.User); err != nil {
			return nil, err
		}

		i := strings.LastIndex(this.User, "_")
		info.EmailSuffix = this.User[(i + 1):]
		info.ID = this.User[:i]
	}
	return dbmodels.GetDB().CheckCorporationManagerExist(info)
}

func CreateCorporationAdministrator(orgCLAID, name, email string) ([]dbmodels.CorporationManagerCreateOption, error) {
	pw := util.RandStr(8, "alphanum")

	opt := []dbmodels.CorporationManagerCreateOption{
		{
			ID:       "admin",
			Name:     name,
			Email:    email,
			Password: pw,
			Role:     dbmodels.RoleAdmin,
		},
	}
	err := dbmodels.GetDB().AddCorporationManager(orgCLAID, opt, 1)
	if err != nil {
		return nil, err
	}

	opt[0].ID = fmt.Sprintf("admin_%s", util.EmailSuffix(opt[0].Email))
	return opt, nil
}

type CorporationManagerResetPassword dbmodels.CorporationManagerResetPassword

func (this CorporationManagerResetPassword) Validate() (string, error) {
	if this.NewPassword == this.OldPassword {
		return util.ErrInvalidParameter, fmt.Errorf("the new password is same as old one")
	}
	return "", nil
}

func (this CorporationManagerResetPassword) Reset(orgCLAID, email string) error {
	return dbmodels.GetDB().ResetCorporationManagerPassword(
		orgCLAID, email, dbmodels.CorporationManagerResetPassword(this),
	)
}

func ListCorporationManagers(linkID, email, role string) ([]dbmodels.CorporationManagerListResult, *ModelError) {
	v, err := dbmodels.GetDB().ListCorporationManager(linkID, email, role)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}

	if err.IsErrorOf(dbmodels.ErrNoChildElem) {
		return v, newModelError(ErrNoCorp, err)
	}

	return v, parseDBError(err)
}
