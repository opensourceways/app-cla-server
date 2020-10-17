package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type CorporationManagerAuthentication dbmodels.CorporationManagerCheckInfo

func (this CorporationManagerAuthentication) Authenticate() (map[string]dbmodels.CorporationManagerCheckResult, error) {
	return dbmodels.GetDB().CheckCorporationManagerExist(
		dbmodels.CorporationManagerCheckInfo(this),
	)
}

func CreateCorporationAdministrator(claOrgID, email string) ([]dbmodels.CorporationManagerCreateOption, error) {
	pw := util.RandStr(8, "alphanum")

	opt := []dbmodels.CorporationManagerCreateOption{
		{
			Role:     dbmodels.RoleAdmin,
			Email:    email,
			Password: pw,
		},
	}
	return dbmodels.GetDB().AddCorporationManager(claOrgID, opt, 1)
}

type CorporationManagerResetPassword dbmodels.CorporationManagerResetPassword

func (this CorporationManagerResetPassword) Validate() (string, error) {
	if this.NewPassword == this.OldPassword {
		return util.ErrInvalidParameter, fmt.Errorf("the new password is same as old one")
	}
	return "", nil
}

func (this CorporationManagerResetPassword) Reset(claOrgID, email string) error {
	return dbmodels.GetDB().ResetCorporationManagerPassword(
		claOrgID, email, dbmodels.CorporationManagerResetPassword(this),
	)
}

func ListCorporationManagers(claOrgID, email, role string) ([]dbmodels.CorporationManagerListResult, error) {
	return dbmodels.GetDB().ListCorporationManager(claOrgID, email, role)
}
