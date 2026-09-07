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

type CorporationSigningPageSummary struct {
	Total int64                       `json:"total"`
	Data  []CorporationSigningSummary `json:"page_data"`
}

type CorporationSigningSummary struct {
	CorporationSigningBasicInfo

	Id             string `json:"id"`
	AdminAdded     bool   `json:"admin_added"`
	PDFUploaded    bool   `json:"pdf_uploaded"`
	AdminAddedDate string `json:"admin_added_date,omitempty"`
}

type CorporationSigningBasicInfo struct {
	CLAId           string `json:"cla_id"`
	CLALanguage     string `json:"cla_language"`
	AdminEmail      string `json:"admin_email"`
	AdminName       string `json:"admin_name"`
	CorporationName string `json:"corporation_name"`
	Date            string `json:"date"`
}

// models/corp_signing.go
type RepresentativeUpdateOption struct {
	RepName  string `json:"rep_name" valid:"Required"`
	RepEmail string `json:"rep_email" valid:"Required;Email"`
}

func UpdateCorpRepresentative(userId, linkID, signingID string, opt *RepresentativeUpdateOption) IModelError {
	return corpSigningAdapterInstance.UpdateRepresentative(userId, linkID, signingID, opt)
}

type CorporationSigningPendingItem struct {
	SigningId         string `json:"signing_id"`
	CorpName          string `json:"corp_name"`
	AdminEmail        string `json:"admin_email"`
	SignedCLAVersion  string `json:"signed_cla_version"`
	PendingCLAVersion string `json:"pending_cla_version"`
	NotifyCount       int    `json:"notify_count"`
	LastNotifyTime    int64  `json:"last_notify_time"`
}

func FindPendingAgreements(userId, linkId string) ([]CorporationSigningPendingItem, IModelError) {
	return corpSigningAdapterInstance.FindPendingAgreements(userId, linkId)
}
