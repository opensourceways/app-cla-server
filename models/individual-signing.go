package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type IndividualSigning dbmodels.IndividualSigningInfo

func (this *IndividualSigning) Validate(email string) *ModelError {
	if this.Email != email {
		return newModelError(ErrUnmatchedEmail, fmt.Errorf("unmatched email"))
	}
	return checkEmailFormat(this.Email)
}

func (this *IndividualSigning) Create(linkID string, enabled bool) *ModelError {
	this.Date = util.Date()
	this.Enabled = enabled

	err := dbmodels.GetDB().SignAsIndividual(
		linkID, (*dbmodels.IndividualSigningInfo)(this),
	)

	if err != nil {
		if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
			return newModelError(ErrNoLinkOrResign, err.Err)
		}
		return parseDBError(err)
	}
	return nil
}

func IsIndividualSigned(orgRepo *dbmodels.OrgRepo, email string) (bool, *ModelError) {
	b, err := dbmodels.GetDB().IsIndividualSigned(orgRepo, email)
	if err != nil {
		return b, parseDBError(err)
	}
	return b, nil
}

func InitializeIndividualSigning(linkID string, orgRepo *dbmodels.OrgRepo, claInfo *dbmodels.CLAInfo) error {
	return dbmodels.GetDB().InitializeIndividualSigning(linkID, orgRepo, claInfo)
}
