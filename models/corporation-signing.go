package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

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

	if _, err := checkEmailFormat(this.AdminEmail); err != nil {
		return newModelError(ErrNotAnEmail, err)
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

func GetCorpSigningDetail(linkID, email string) (*dbmodels.CorpSigningCreateOpt, IModelError) {
	s, err := dbmodels.GetDB().GetCorpSigningDetail(linkID, email)
	if err == nil {
		if s == nil {
			return nil, newModelError(ErrUnsigned, fmt.Errorf("unsigned"))
		}
		return s, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return s, newModelError(ErrNoLink, err)
	}

	return s, parseDBError(err)
}
