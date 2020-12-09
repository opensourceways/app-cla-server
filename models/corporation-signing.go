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

func (this *CorporationSigningCreateOption) Create(orgCLAID, platform, orgID, repoId string) error {
	this.Date = util.Date()

	return dbmodels.GetDB().SignAsCorporation(
		orgCLAID, platform, orgID, repoId,
		dbmodels.CorporationSigningInfo(this.CorporationSigning),
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

func GetCorporationSigningDetail(platform, org, repo, email string) (string, dbmodels.CorporationSigningDetail, error) {
	return dbmodels.GetDB().GetCorporationSigningDetail(platform, org, repo, email)
}

func GetCorpSigningInfo(platform, org, repo, email string) (string, *dbmodels.CorporationSigningInfo, error) {
	return dbmodels.GetDB().GetCorpSigningInfo(platform, org, repo, email)
}

type CorporationSigningListOption dbmodels.CorporationSigningListOption

func (this CorporationSigningListOption) List() (map[string][]dbmodels.CorporationSigningDetail, error) {
	return dbmodels.GetDB().ListCorporationSigning(dbmodels.CorporationSigningListOption(this))
}
