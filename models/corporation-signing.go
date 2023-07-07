package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type CorporationSigning = dbmodels.CorpSigningCreateOpt

type CorporationSigningCreateOption struct {
	CorporationSigning

	VerificationCode string `json:"verification_code"`
}

func (this *CorporationSigningCreateOption) Validate(linkId string) IModelError {
	if err := checkEmailFormat(this.AdminEmail); err != nil {
		return err
	}

	err := validateCodeForSigning(linkId, this.AdminEmail, this.VerificationCode)
	if err != nil {
		return err
	}

	return nil
}

type CorporationSigningSummary struct {
	dbmodels.CorporationSigningBasicInfo
	Id          string `json:"string"`
	AdminAdded  bool   `json:"admin_added"`
	PDFUploaded bool   `json:"pdf_uploaded"`
}
