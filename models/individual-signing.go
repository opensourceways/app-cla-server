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

func (this *IndividualSigning) Create(orgCLAID string, enabled bool) error {
	this.Date = util.Date()
	this.Enabled = enabled

	return dbmodels.GetDB().SignAsIndividual(
		orgCLAID, (*dbmodels.IndividualSigningInfo)(this),
	)
}

func IsIndividualSigned(orgRepo *dbmodels.OrgRepo, email string) (bool, error) {
	return dbmodels.GetDB().IsIndividualSigned(orgRepo, email)
}

func GetCLAInfoSigned(linkID, claLang, applyTo string) (*dbmodels.CLAInfo, error) {
	return dbmodels.GetDB().GetCLAInfoSigned(linkID, claLang, applyTo)
}

func GetCLAInfoToSign(linkID, claLang, applyTo string) (*dbmodels.CLAInfo, error) {
	return dbmodels.GetDB().GetCLAInfoToSign(linkID, claLang, applyTo)
}
