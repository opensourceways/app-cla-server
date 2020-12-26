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

func DownloadCorporationSigningPDF(linkID string, email string) (*[]byte, *ModelError) {
	v, err := dbmodels.GetDB().DownloadCorporationSigningPDF(linkID, email)
	return v, parseDBError(err)
}

func IsCorpSigningPDFUploaded(linkID string, email string) (bool, *ModelError) {
	v, err := dbmodels.GetDB().IsCorpSigningPDFUploaded(linkID, email)
	return v, parseDBError(err)
}

func ListCorpsWithPDFUploaded(linkID string) ([]string, *ModelError) {
	v, err := dbmodels.GetDB().ListCorpsWithPDFUploaded(linkID)
	return v, parseDBError(err)
}

func GetCorporationSigningBasicInfo(linkID, email string) (*dbmodels.CorporationSigningBasicInfo, *ModelError) {
	v, err := dbmodels.GetDB().GetCorpSigningBasicInfo(linkID, email)
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

func GetCorpSigningDetail(linkID, email string) ([]dbmodels.Field, *dbmodels.CorporationSigningOption, *ModelError) {
	f, s, err := dbmodels.GetDB().GetCorpSigningDetail(linkID, email)
	if err == nil {
		return f, s, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return f, s, newModelError(ErrNoLink, err)
	}

	if err.IsErrorOf(dbmodels.ErrNoChildElem) {
		return f, s, newModelError(ErrUnsigned, err)
	}

	return f, s, parseDBError(err)
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

func IsCorpSigned(linkID, email string) (bool, *ModelError) {
	v, err := dbmodels.GetDB().IsCorpSigned(linkID, email)
	if err == nil {
		return v, nil
	}

	return v, parseDBError(err)
}
