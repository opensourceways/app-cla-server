package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type IndividualSigning struct {
	Email            string                   `json:"email"`
	Name             string                   `json:"name"`
	CLAId            string                   `json:"cla_id"`
	CLALanguage      string                   `json:"cla_language"`
	VerificationCode string                   `json:"verification_code"`
	Info             dbmodels.TypeSigningInfo `json:"info"`
}
