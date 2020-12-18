package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type EmployeeSigning struct {
	IndividualSigning

	VerificationCode string `json:"verification_code"`
}

func (this *EmployeeSigning) Validate(orgCLAID, email string) *ModelError {
	err := checkVerificationCode(this.Email, this.VerificationCode, orgCLAID)
	if err != nil {
		return err
	}

	return (&this.IndividualSigning).Validate(email)
}

type EmployeeSigningListOption struct {
	CLALanguage string `json:"cla_language"`
}

func (this EmployeeSigningListOption) List(linkID, corpEmail string) ([]dbmodels.IndividualSigningBasicInfo, *ModelError) {
	v, err := dbmodels.GetDB().ListIndividualSigning(linkID, corpEmail, this.CLALanguage)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return nil, newModelError(ErrNoLink, err.Err)
	}
	return nil, parseDBError(err)
}

type EmployeeSigningUdateInfo struct {
	Enabled bool `json:"enabled"`
}

func (this *EmployeeSigningUdateInfo) Update(linkID, email string) *ModelError {
	err := dbmodels.GetDB().UpdateIndividualSigning(linkID, email, this.Enabled)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrUnsigned, err)
	}
	return parseDBError(err)
}

func DeleteEmployeeSigning(linkID, email string) *ModelError {
	err := dbmodels.GetDB().DeleteIndividualSigning(linkID, email)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLink, err)
	}
	return parseDBError(err)
}
