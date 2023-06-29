package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/config"
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

func (this *CorporationSigningCreateOption) Validate(linkId string) IModelError {
	if err := checkEmailFormat(this.AdminEmail); err != nil {
		return err
	}

	err := validateCodeForSigning(linkId, this.AdminEmail, this.VerificationCode)
	if err != nil {
		return err
	}

	if config.AppConfig.IsRestrictedEmailSuffix(util.EmailSuffix(this.AdminEmail)) {
		return newModelError(ErrRestrictedEmailSuffix, fmt.Errorf("email suffix is restricted"))
	}
	return nil
}

func (this *CorporationSigningCreateOption) Create(linkId string) IModelError {
	if corpSigningAdapterInstance != nil {
		return corpSigningAdapterInstance.Sign(this, linkId)
	}

	this.Date = util.Date()

	err := dbmodels.GetDB().SignCorpCLA(linkId, &this.CorporationSigning)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrResigned, err)
	}
	return parseDBError(err)
}

func UploadCorporationSigningPDF(linkID string, email string, pdf []byte) IModelError {
	err := dbmodels.GetDB().UploadCorporationSigningPDF(linkID, email, pdf)
	return parseDBError(err)
}

func DownloadCorporationSigningPDF(linkID string, email string) ([]byte, IModelError) {
	v, err := dbmodels.GetDB().DownloadCorporationSigningPDF(linkID, email)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLinkOrUnuploaed, err)
	}
	return v, parseDBError(err)
}

func IsCorpSigningPDFUploaded(linkID string, email string) (bool, IModelError) {
	v, err := dbmodels.GetDB().IsCorpSigningPDFUploaded(linkID, email)
	return v, parseDBError(err)
}

func ListCorpsWithPDFUploaded(linkID string) ([]string, IModelError) {
	v, err := dbmodels.GetDB().ListCorpsWithPDFUploaded(linkID)
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

func GetCorpSigningDetail(linkID, email string) (*dbmodels.CLAInfo, *dbmodels.CorpSigningCreateOpt, IModelError) {
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

func RemoveCorpSigning(csId string) IModelError {
	return corpSigningAdapterInstance.Remove(csId)
}

type CorporationSigningSummary struct {
	dbmodels.CorporationSigningBasicInfo
	Id          string `json:"string"`
	AdminAdded  bool   `json:"admin_added"`
	PDFUploaded bool   `json:"pdf_uploaded"`
}

func ListCorpSigning(linkID string) ([]CorporationSigningSummary, IModelError) {
	return corpSigningAdapterInstance.List(linkID)
}

func GetCorpSigning(csId string) (CorporationSigning, IModelError) {
	return corpSigningAdapterInstance.Get(csId)
}

func UploadCorpPDF(csId string, pdf []byte) IModelError {
	return corpPDFAdapterInstance.Upload(csId, pdf)
}

func DownloadCorpPDF(csId string) ([]byte, IModelError) {
	return corpPDFAdapterInstance.Download(csId)
}
