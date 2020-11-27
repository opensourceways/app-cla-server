package models

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type CorporationSigning dbmodels.CorporationSigningInfo

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

func (this *CorporationSigningCreateOption) Create(orgRepo *dbmodels.OrgRepo) error {
	this.Date = util.Date()

	return dbmodels.GetDB().SignAsCorporation(
		orgRepo,
		(*dbmodels.CorporationSigningInfo)(&this.CorporationSigning),
	)
}

func UploadCorporationSigningPDF(orgRepo *dbmodels.OrgRepo, email string, pdf []byte) error {
	return dbmodels.GetDB().UploadCorporationSigningPDF(orgRepo, email, pdf)
}

func DownloadCorporationSigningPDF(orgRepo *dbmodels.OrgRepo, email string) ([]byte, error) {
	return dbmodels.GetDB().DownloadCorporationSigningPDF(orgRepo, email)
}

func GetCorporationSigningDetail(orgRepo *dbmodels.OrgRepo, email string) (dbmodels.CorporationSigningDetail, error) {
	return dbmodels.GetDB().GetCorporationSigningDetail(orgRepo, email)
}

func GetCorporationSigningSummary(orgRepo *dbmodels.OrgRepo, email string) (dbmodels.CorporationSigningSummary, error) {
	return dbmodels.GetDB().GetCorporationSigningSummary(orgRepo, email)
}

func ListCorporationSigning(orgRepo *dbmodels.OrgRepo, language string) ([]dbmodels.CorporationSigningSummary, error) {
	return dbmodels.GetDB().ListCorporationSigning(orgRepo, language)
}
