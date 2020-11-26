package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type IndividualSigning dbmodels.IndividualSigningInfo

func (this *IndividualSigning) Validate(email string) (string, error) {
	if this.Email != email {
		return util.ErrInvalidParameter, fmt.Errorf("not authorized email")
	}
	return checkEmailFormat(this.Email)
}

func (this *IndividualSigning) Create(orgRepo *dbmodels.OrgRepo, enabled bool) error {
	this.Date = util.Date()
	this.Enabled = enabled

	return dbmodels.GetDB().SignAsIndividual(
		orgRepo, (*dbmodels.IndividualSigningInfo)(this),
	)
}

func IsIndividualSigned(orgRepo *dbmodels.OrgRepo, email string) (bool, error) {
	return dbmodels.GetDB().IsIndividualSigned(orgRepo, email)
}
