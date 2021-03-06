package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

func InitializeCorpSigning(linkID string, info *OrgInfo, cla *CLAInfo) IModelError {
	err := dbmodels.GetDB().InitializeCorpSigning(linkID, info, cla)
	return parseDBError(err)
}

type CorporationSigning = dbmodels.CorpSigningCreateOpt

type CorporationSigningCreateOption struct {
	CorporationSigning

	VerificationCode string `json:"verification_code"`
}

func (this *CorporationSigningCreateOption) Validate(orgCLAID string) IModelError {
	err := checkVerificationCode(this.AdminEmail, this.VerificationCode, orgCLAID)
	if err != nil {
		return err
	}
	return checkEmailFormat(this.AdminEmail)
}

func (this *CorporationSigningCreateOption) Create(orgCLAID string) IModelError {
	this.Date = util.Date()

	err := dbmodels.GetDB().SignCorpCLA(orgCLAID, &this.CorporationSigning)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrResigned, err)
	}
	return parseDBError(err)
}

func UploadCorporationSigningPDF(linkID, email string, pdf []byte) IModelError {
	err := dbmodels.GetDB().UploadCorporationSigningPDF(linkID, email, pdf)
	return parseDBError(err)
}

func DownloadCorporationSigningPDF(linkID, email, path string) IModelError {
	err := dbmodels.GetDB().DownloadCorporationSigningPDF(linkID, email, path)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrUnuploaed, err)
	}
	return parseDBError(err)
}

func IsCorpSigningPDFUploaded(linkID string, email string) (bool, IModelError) {
	v, err := dbmodels.GetDB().IsCorporationSigningPDFUploaded(linkID, email)
	return v, parseDBError(err)
}

func ListCorpsWithPDFUploaded(linkID string) ([]string, IModelError) {
	v, err := dbmodels.GetDB().ListCorporationsWithPDFUploaded(linkID)
	return v, parseDBError(err)
}

func ListCorpSignings(linkID, language string) ([]dbmodels.CorporationSigningSummary, IModelError) {
	v, err := dbmodels.GetDB().ListCorpSignings(linkID, language)
	if err == nil {
		if v == nil {
			v = []dbmodels.CorporationSigningSummary{}
		}
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}
	return v, parseDBError(err)
}

func IsCorpSigned(linkID, email string) (bool, IModelError) {
	v, err := dbmodels.GetDB().IsCorpSigned(linkID, email)
	if err == nil {
		return v, nil
	}

	return v, parseDBError(err)
}

func GetCorpSigningBasicInfo(linkID, email string) (*dbmodels.CorporationSigningBasicInfo, IModelError) {
	v, err := dbmodels.GetDB().GetCorpSigningBasicInfo(linkID, email)
	if err == nil {
		if v == nil {
			return nil, newModelError(ErrUnsigned, fmt.Errorf("unsigned"))
		}
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}

	return v, parseDBError(err)
}

func GetCorpSigningDetail(linkID, email string) ([]dbmodels.Field, *dbmodels.CorpSigningCreateOpt, IModelError) {
	f, s, err := dbmodels.GetDB().GetCorpSigningDetail(linkID, email)
	if err == nil {
		return f, s, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return f, s, newModelError(ErrNoLink, err)
	}

	return f, s, parseDBError(err)
}

func DeleteCorpSigning(linkID, email string) IModelError {
	err := dbmodels.GetDB().DeleteCorpSigning(linkID, email)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLink, err)
	}
	return parseDBError(err)
}

func ListDeletedCorpSignings(linkID string) ([]dbmodels.CorporationSigningBasicInfo, IModelError) {
	v, err := dbmodels.GetDB().ListDeletedCorpSignings(linkID)
	if err == nil {
		if v == nil {
			v = []dbmodels.CorporationSigningBasicInfo{}
		}
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}
	return v, parseDBError(err)
}
