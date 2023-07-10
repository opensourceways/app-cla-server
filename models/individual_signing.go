package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type IndividualSigning struct {
	dbmodels.IndividualSigningInfo

	VerificationCode string `json:"verification_code"`
}
