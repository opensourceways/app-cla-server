package models

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type EmployeeSigning struct {
	IndividualSigning

	CorpSigningId string `json:"corp_signing_id"`
}

func (e *EmployeeSigning) Create(linkID string) IModelError {
	e.Date = util.Date()

	err := dbmodels.GetDB().SignEmployeeCLA(
		&SigningIndex{
			LinkId:    linkID,
			SigningId: e.CorpSigningId,
		},
		&e.IndividualSigningInfo,
	)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrResigned, err)
	}

	return parseDBError(err)
}

func ListEmployeeSigning(index SigningIndex, claLang string) (
	[]dbmodels.IndividualSigningBasicInfo, IModelError,
) {
	v, err := dbmodels.GetDB().ListEmployeeSigning(&index, claLang)
	if err == nil {
		return v, nil
	}

	return nil, parseDBError(err)
}

type EmployeeSigningUdateInfo struct {
	Enabled bool `json:"enabled"`
}

func (this *EmployeeSigningUdateInfo) Update(index SigningIndex) (string, IModelError) {
	v, err := dbmodels.GetDB().UpdateEmployeeSigning(&index, this.Enabled)
	if err == nil {
		return v.Email, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return "", newModelError(ErrNoLinkOrUnsigned, err)
	}
	return "", parseDBError(err)
}

func DeleteEmployeeSigning(index SigningIndex) (string, IModelError) {
	v, err := dbmodels.GetDB().DeleteIndividualSigning(&index)
	if err == nil {
		return v.Email, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return "", newModelError(ErrNoLink, err)
	}
	return "", parseDBError(err)
}
