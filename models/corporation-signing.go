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

func (this *CorporationSigningCreateOption) Validate() (string, error) {
	ec, err := checkVerificationCode(this.AdminEmail, this.VerificationCode, ActionCorporationSigning)
	if err != nil {
		return ec, err
	}

	return checkEmailFormat(this.AdminEmail)
}

func (this *CorporationSigningCreateOption) Create(orgCLAID, platform, orgID, repoId string) error {
	this.Date = util.Date()

	return dbmodels.GetDB().SignAsCorporation(
		orgCLAID, platform, orgID, repoId,
		dbmodels.CorporationSigningInfo(this.CorporationSigning),
	)
}

func CheckCorporationSigning(orgCLAID, email string) (dbmodels.CorporationSigningDetail, error) {
	return dbmodels.GetDB().CheckCorporationSigning(orgCLAID, email)
}

func UploadCorporationSigningPDF(orgCLAID, email string, pdf []byte) error {
	return dbmodels.GetDB().UploadCorporationSigningPDF(orgCLAID, email, pdf)
}

func DownloadCorporationSigningPDF(orgCLAID, email string) ([]byte, error) {
	return dbmodels.GetDB().DownloadCorporationSigningPDF(orgCLAID, email)
}

func GetCorporationSigningDetail(platform, org, repo, email string) (string, dbmodels.CorporationSigningDetail, error) {
	return dbmodels.GetDB().GetCorporationSigningDetail(platform, org, repo, email)
}

type CorporationSigningListOption dbmodels.CorporationSigningListOption

func (this CorporationSigningListOption) List() (map[string][]dbmodels.CorporationSigningDetail, error) {
	return dbmodels.GetDB().ListCorporationSigning(dbmodels.CorporationSigningListOption(this))
}
