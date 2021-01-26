package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type EmployeeSigning struct {
	IndividualSigning

	VerificationCode string `json:"verification_code"`
}

func (this *EmployeeSigning) Validate(linkID, userID, email string) IModelError {
	if err := checkVerificationCode(this.Email, this.VerificationCode, linkID); err != nil {
		return err
	}

	return (&this.IndividualSigning).Validate(userID, email)
}

func ListIndividualSigning(linkID, corpEmail, claLang string) ([]dbmodels.IndividualSigningBasicInfo, IModelError) {
	v, err := dbmodels.GetDB().ListIndividualSigning(linkID, corpEmail, claLang)
	if err == nil {
		return v, nil
	}

	return nil, parseDBError(err)
}

type EmployeeSigningUdateInfo struct {
	Enabled bool `json:"enabled"`
}

func (this *EmployeeSigningUdateInfo) Update(linkID, email string) IModelError {
	err := dbmodels.GetDB().UpdateIndividualSigning(linkID, email, this.Enabled)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLinkOrUnsigned, err)
	}
	return parseDBError(err)
}

func DeleteEmployeeSigning(linkID, email string) IModelError {
	err := dbmodels.GetDB().DeleteIndividualSigning(linkID, email)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLink, err)
	}
	return parseDBError(err)
}
