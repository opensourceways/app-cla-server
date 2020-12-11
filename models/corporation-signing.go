package models

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type CorporationSigning dbmodels.CorporationSigningOption

type CorporationSigningCreateOption struct {
	CorporationSigning

	VerificationCode string `json:"verification_code"`
}

func (this *CorporationSigningCreateOption) Validate(orgCLAID string) (string, error) {
	ec, err := checkVerificationCode(this.AdminEmail, this.VerificationCode, orgCLAID)
	if err != nil {
		return ec, err
	}

	return checkEmailFormat(this.AdminEmail)
}

func (this *CorporationSigningCreateOption) Create(orgCLAID string) error {
	this.Date = util.Date()

	return dbmodels.GetDB().SignAsCorporation(
		orgCLAID,
		(*dbmodels.CorporationSigningOption)(&this.CorporationSigning),
	)
}

func UploadCorporationSigningPDF(linkID string, email string, pdf *[]byte) error {
	return dbmodels.GetDB().UploadCorporationSigningPDF(linkID, email, pdf)
}

func DownloadCorporationSigningPDF(linkID string, email string) (*[]byte, error) {
	return dbmodels.GetDB().DownloadCorporationSigningPDF(linkID, email)
}

func IsCorpSigningPDFUploaded(linkID string, email string) (bool, error) {
	return dbmodels.GetDB().IsCorpSigningPDFUploaded(linkID, email)
}

func ListCorpsWithPDFUploaded(linkID string) ([]string, error) {
	return dbmodels.GetDB().ListCorpsWithPDFUploaded(linkID)
}

func GetCorporationSigningBasicInfo(linkID, email string) (*dbmodels.CorporationSigningBasicInfo, error) {
	return dbmodels.GetDB().GetCorpSigningBasicInfo(linkID, email)
}

func GetCorpSigningDetail(linkID, email string) (*dbmodels.CorporationSigningOption, error) {
	return dbmodels.GetDB().GetCorpSigningDetail(linkID, email)
}

func ListCorpSignings(linkID, language string) ([]dbmodels.CorporationSigningSummary, error) {
	return dbmodels.GetDB().ListCorpSignings(linkID, language)
}
