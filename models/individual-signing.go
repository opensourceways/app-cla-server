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

func (this *IndividualSigning) Create(platform, orgID, repoId string, enabled bool) error {
	this.Date = util.Date()
	this.Enabled = enabled

	return dbmodels.GetDB().SignAsIndividual(
		platform, orgID, repoId, *(*dbmodels.IndividualSigningInfo)(this),
	)
}

func IsIndividualSigned(platform, orgID, repoId, email string) (bool, error) {
	return dbmodels.GetDB().IsIndividualSigned(platform, orgID, repoId, email)
}
