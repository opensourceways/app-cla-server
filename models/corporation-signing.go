package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type SigningIndex = dbmodels.SigningIndex

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
	if err := checkEmailFormat(this.AdminEmail); err != nil {
		return err
	}

	if err := checkVerificationCode(this.AdminEmail, this.VerificationCode, orgCLAID); err != nil {
		return err
	}

	if config.AppConfig.IsRestrictedEmailSuffix(util.EmailSuffix(this.AdminEmail)) {
		return newModelError(ErrRestrictedEmailSuffix, fmt.Errorf("email suffix is restricted"))
	}
	return nil
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

func UploadCorporationSigningPDF(index SigningIndex, pdf []byte) IModelError {
	err := dbmodels.GetDB().UploadCorporationSigningPDF(&index, pdf)
	return parseDBError(err)
}

func DownloadCorporationSigningPDF(index *SigningIndex) ([]byte, IModelError) {
	v, err := dbmodels.GetDB().DownloadCorporationSigningPDF(index)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLinkOrUnuploaed, err)
	}
	return v, parseDBError(err)
}

func IsCorpSigningPDFUploaded(index SigningIndex) (bool, IModelError) {
	v, err := dbmodels.GetDB().IsCorpSigningPDFUploaded(&index)
	return v, parseDBError(err)
}

func ListCorpsWithPDFUploaded(linkID string) ([]string, IModelError) {
	v, err := dbmodels.GetDB().ListCorpsWithPDFUploaded(linkID)
	return v, parseDBError(err)
}

func ListCorpSignings(linkID string, opt dbmodels.CorpSigningListOpt) (
	[]dbmodels.CorporationSigningSummary, IModelError,
) {
	v, err := dbmodels.GetDB().ListCorpSignings(linkID, &opt)
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
	v, err := dbmodels.GetDB().ListCorpSignings(linkID, &dbmodels.CorpSigningListOpt{
		Email: email,
	})
	if err == nil {
		return len(v) > 0, nil
	}

	return false, parseDBError(err)
}

func GetCorpSigningBasicInfo(index *SigningIndex) (*dbmodels.CorporationSigningBasicInfo, IModelError) {
	v, err := dbmodels.GetDB().GetCorpSigningBasicInfo(index)
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

func GetCorpSigningDetail(index SigningIndex) (*dbmodels.CLAInfo, *dbmodels.CorpSigningCreateOpt, IModelError) {
	f, s, err := dbmodels.GetDB().GetCorpSigningDetail(&index)
	if err == nil {
		return f, s, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return f, s, newModelError(ErrNoLink, err)
	}

	return f, s, parseDBError(err)
}

func DeleteCorpSigning(index SigningIndex) IModelError {
	err := dbmodels.GetDB().DeleteCorpSigning(&index)
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
