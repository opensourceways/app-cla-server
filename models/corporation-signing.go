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

func (this *CorporationSigningCreateOption) Validate(orgCLAID string) *ModelError {
	err := checkVerificationCode(this.AdminEmail, this.VerificationCode, orgCLAID)
	if err != nil {
		return err
	}

	return checkEmailFormat(this.AdminEmail)
}

func (this *CorporationSigningCreateOption) Create(orgCLAID string) *ModelError {
	this.Date = util.Date()

	err := dbmodels.GetDB().SignAsCorporation(
		orgCLAID,
		(*dbmodels.CorporationSigningOption)(&this.CorporationSigning),
	)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrResign, err.Err)
	}
	return parseDBError(err)

}

func UploadCorporationSigningPDF(linkID string, email string, pdf *[]byte) *ModelError {
	err := dbmodels.GetDB().UploadCorporationSigningPDF(linkID, email, pdf)
	return parseDBError(err)
}

func DownloadCorporationSigningPDF(linkID string, email string) (*[]byte, error) {
	return dbmodels.GetDB().DownloadCorporationSigningPDF(linkID, email)
}

func IsCorpSigningPDFUploaded(linkID string, email string) (bool, error) {
	return dbmodels.GetDB().IsCorpSigningPDFUploaded(linkID, email)
}

func ListCorpsWithPDFUploaded(linkID string) ([]string, *ModelError) {
	v, err := dbmodels.GetDB().ListCorpsWithPDFUploaded(linkID)
	return v, parseDBError(err)
}

func GetCorporationSigningBasicInfo(linkID, email string) (*dbmodels.CorporationSigningBasicInfo, error) {
	return dbmodels.GetDB().GetCorpSigningBasicInfo(linkID, email)
}

func GetCorpSigningDetail(linkID, email string) (*dbmodels.CorporationSigningOption, error) {
	return dbmodels.GetDB().GetCorpSigningDetail(linkID, email)
}

func ListCorpSignings(linkID, language string) ([]dbmodels.CorporationSigningSummary, *ModelError) {
	v, err := dbmodels.GetDB().ListCorpSignings(linkID, language)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}

	if err.IsErrorOf(dbmodels.ErrNoChildElem) {
		return v, newModelError(ErrNoCorp, err)
	}

	return v, parseDBError(err)
}

func InitializeCorpSigning(linkID string, info *dbmodels.OrgInfo, claInfo *dbmodels.CLAInfo) error {
	return dbmodels.GetDB().InitializeCorpSigning(linkID, info, claInfo)
}

func IsCorpSigned(linkID, email string) (bool, *ModelError) {
	v, err := dbmodels.GetDB().IsCorpSigned(linkID, email)
	if err == nil {
		return v, nil
	}

	return v, parseDBError(err)
}
