package models

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func InitializeIndividualSigning(linkID string, cla *CLAInfo) IModelError {
	err := dbmodels.GetDB().InitializeIndividualSigning(linkID, cla)
	return parseDBError(err)
}

type IndividualSigning struct {
	dbmodels.IndividualSigningInfo

	VerificationCode string `json:"verification_code"`
}

func (isign *IndividualSigning) Validate(linkID string) IModelError {
	return checkVerificationCode(isign.Email, isign.VerificationCode, linkID)
}

func (isign *IndividualSigning) Create(linkID string, enabled bool) IModelError {
	isign.Date = util.Date()
	isign.Enabled = enabled

	err := dbmodels.GetDB().SignIndividualCLA(
		linkID, &isign.IndividualSigningInfo,
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
