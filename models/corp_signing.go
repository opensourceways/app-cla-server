package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type CorporationSigning = dbmodels.CorpSigningCreateOpt

type CorporationSigningCreateOption struct {
	CorporationSigning

	VerificationCode string `json:"verification_code"`
}

type CorporationSigningSummary struct {
	dbmodels.CorporationSigningBasicInfo
	Id          string `json:"string"`
	AdminAdded  bool   `json:"admin_added"`
	PDFUploaded bool   `json:"pdf_uploaded"`
}
