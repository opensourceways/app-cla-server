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

func (this *CorporationSigningCreateOption) Validate() error {
	return checkVerificationCode(this.AdminEmail, this.VerificationCode, ActionCorporationSigning)
}

func (this *CorporationSigningCreateOption) Create(claOrgID, platform, orgID, repoId string) error {
	this.Date = util.Date()

	return dbmodels.GetDB().SignAsCorporation(
		claOrgID, platform, orgID, repoId,
		dbmodels.CorporationSigningInfo(this.CorporationSigning),
	)
}

func CheckCorporationSigning(claOrgID, email string) (dbmodels.CorporationSigningDetail, error) {
	return dbmodels.GetDB().CheckCorporationSigning(claOrgID, email)
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
