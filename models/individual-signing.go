package models

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type IndividualSigning dbmodels.IndividualSigningInfo

func (this *IndividualSigning) Validate() (string, error) {
	return checkEmailFormat(this.Email)
}

func (this *IndividualSigning) Create(orgCLAID, platform, orgID, repoId string, enabled bool) error {
	this.Date = util.Date()
	this.Enabled = enabled

	return dbmodels.GetDB().SignAsIndividual(
		orgCLAID, platform, orgID, repoId, *(*dbmodels.IndividualSigningInfo)(this),
	)
}

func IsIndividualSigned(platform, orgID, repoId, email string) (bool, error) {
	return dbmodels.GetDB().IsIndividualSigned(platform, orgID, repoId, email)
}
