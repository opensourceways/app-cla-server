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

func (this EmployeeSigningListOption) List(linkID, corpEmail string) ([]dbmodels.IndividualSigningBasicInfo, error) {
	return dbmodels.GetDB().ListIndividualSigning(linkID, corpEmail, this.CLALanguage)
}

type EmployeeSigningUdateInfo struct {
	Enabled bool `json:"enabled"`
}

func (this *EmployeeSigningUdateInfo) Update(linkID, email string) error {
	return dbmodels.GetDB().UpdateIndividualSigning(linkID, email, this.Enabled)
}

func DeleteEmployeeSigning(linkID, email string) error {
	return dbmodels.GetDB().DeleteIndividualSigning(linkID, email)
}
