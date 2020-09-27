package models

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

const ActionCorporationSigning = "corporation-signing"

type CorporationSigning dbmodels.CorporationSigningInfo

type CorporationSigningCreateOption struct {
	CorporationSigning

	VerifiCode string `json:"verifi_code"`
}

func (this *CorporationSigningCreateOption) Validate() error {
	vc := dbmodels.VerificationCode{
		Email:   this.AdminEmail,
		Code:    this.VerifiCode,
		Purpose: ActionCorporationSigning,
	}

	return dbmodels.GetDB().CheckVerificationCode(vc)
}

func (this *CorporationSigningCreateOption) Create(claOrgID string) error {
	this.Date = util.Date()

	return dbmodels.GetDB().SignAsCorporation(claOrgID, dbmodels.CorporationSigningInfo(this.CorporationSigning))
}

func UploadCorporationSigningPDF(claOrgID, email string, pdf []byte) error {
	return dbmodels.GetDB().UploadCorporationSigningPDF(claOrgID, email, pdf)
}

func DownloadCorporationSigningPDF(claOrgID, email string) ([]byte, error) {
	return dbmodels.GetDB().DownloadCorporationSigningPDF(claOrgID, email)
}

func GetCorporationSigningDetail(platform, org, repo, email string) (string, dbmodels.CorporationSigningDetail, error) {
	return dbmodels.GetDB().GetCorporationSigningDetail(platform, org, repo, email)
}

type CorporationSigningListOption dbmodels.CorporationSigningListOption

func (this CorporationSigningListOption) List() (map[string][]dbmodels.CorporationSigningDetail, error) {
	return dbmodels.GetDB().ListCorporationSigning(dbmodels.CorporationSigningListOption(this))
}

type CorporationSigningVerifCode struct {
	// Email is the email address of corporation
	Email string `json:"email"`
}

func (this CorporationSigningVerifCode) Create(expiry int64) (string, error) {
	code := "123456"

	vc := dbmodels.VerificationCode{
		Email:   this.Email,
		Code:    code,
		Purpose: ActionCorporationSigning,
		Expiry:  util.Now() + expiry,
	}

	err := dbmodels.GetDB().CreateVerificationCode(vc)
	return code, err
}
