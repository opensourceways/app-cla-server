package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type IndividualSigning struct {
	dbmodels.IndividualSigningInfo

	VerificationCode string `json:"verification_code"`
}

func (isign *IndividualSigning) Validate(linkID string) IModelError {
	return validateCodeForSigning(linkID, isign.Email, isign.VerificationCode)
}
