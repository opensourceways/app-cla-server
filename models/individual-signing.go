package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func InitializeIndividualSigning(linkID string, cla *CLAInfo) IModelError {
	err := dbmodels.GetDB().InitializeIndividualSigning(linkID, cla)
	return parseDBError(err)
}

type IndividualSigning dbmodels.IndividualSigningInfo

func (this *IndividualSigning) Validate(email string) IModelError {
	if this.Email != email {
		return newModelError(ErrUnmatchedEmail, fmt.Errorf("unmatched email"))
	}
	return checkEmailFormat(this.Email)
}

func (this *IndividualSigning) Create(linkID string, enabled bool) IModelError {
	this.Date = util.Date()
	this.Enabled = enabled

	err := dbmodels.GetDB().SignIndividualCLA(
		linkID, (*dbmodels.IndividualSigningInfo)(this),
	)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrResigned, err)
	}
	return parseDBError(err)
}

func IsIndividualSigned(linkID, email string) (bool, IModelError) {
	b, err := dbmodels.GetDB().IsIndividualSigned(linkID, email)
	if err == nil {
		return b, nil
	}
	return b, parseDBError(err)
}
