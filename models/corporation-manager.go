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

func (this CorporationManagerAuthentication) Authenticate() ([]dbmodels.CorporationManagerCheckResult, error) {
	info := dbmodels.CorporationManagerCheckInfo{Password: this.Password}
	if _, err := checkEmailFormat(this.User); err == nil {
		info.Email = this.User
	} else {
		if _, err = checkManagerID(this.User); err != nil {
			return nil, err
		}

		i := strings.LastIndex(this.User, "_")
		info.EmailSuffix = this.User[(i + 1):]
		info.ID = this.User[:i]
	}
	return dbmodels.GetDB().GetCorpManager(&info)
}

func CreateCorporationAdministrator(orgRepo *dbmodels.OrgRepo, name, email string) ([]dbmodels.CorporationManagerCreateOption, error) {
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
	err := dbmodels.GetDB().AddCorporationManager(orgRepo, opt, 1)
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

func (this CorporationManagerResetPassword) Reset(orgRepo *dbmodels.OrgRepo, email string) error {
	return dbmodels.GetDB().ResetCorporationManagerPassword(
		orgRepo, email, (*dbmodels.CorporationManagerResetPassword)(&this),
	)
}

func ListCorporationManagers(orgRepo *dbmodels.OrgRepo, email, role string) ([]dbmodels.CorporationManagerListResult, error) {
	return dbmodels.GetDB().ListCorporationManager(orgRepo, email, role)
}
