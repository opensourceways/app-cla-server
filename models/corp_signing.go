package models

import "github.com/opensourceways/app-cla-server/util"

type TypeSigningInfo map[string]string

type CorporationSigning struct {
	CorporationSigningBasicInfo

	Info TypeSigningInfo `json:"info"`
}

type CorporationSigningCreateOption struct {
	CLAId            string          `json:"cla_id"`
	CLALanguage      string          `json:"cla_language"`
	AdminName        string          `json:"admin_name"`
	AdminEmail       string          `json:"admin_email"`
	CorporationName  string          `json:"corporation_name"`
	VerificationCode string          `json:"verification_code"`
	Info             TypeSigningInfo `json:"info"`
	PrivacyChecked   bool            `json:"privacy_checked"`
}

func (opt *CorporationSigningCreateOption) ToCorporationSigning() CorporationSigning {
	return CorporationSigning{
		CorporationSigningBasicInfo: CorporationSigningBasicInfo{
			CLAId:           opt.CLAId,
			CLALanguage:     opt.CLALanguage,
			AdminName:       opt.AdminName,
			AdminEmail:      opt.AdminEmail,
			CorporationName: opt.CorporationName,
			Date:            util.Date(),
		},
		Info: opt.Info,
	}
}

type CorporationSigningSummary struct {
	CorporationSigningBasicInfo

	Id          string `json:"id"`
	AdminAdded  bool   `json:"admin_added"`
	PDFUploaded bool   `json:"pdf_uploaded"`
}

type CorporationSigningBasicInfo struct {
	CLAId           string `json:"cla_id"`
	CLALanguage     string `json:"cla_language"`
	AdminEmail      string `json:"admin_email"`
	AdminName       string `json:"admin_name"`
	CorporationName string `json:"corporation_name"`
	Date            string `json:"date"`
}
