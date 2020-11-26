package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type EmployeeSigning struct {
	IndividualSigning

	VerificationCode string `json:"verification_code"`
}

func (this *EmployeeSigning) Validate(orgCLAID, email string) (string, error) {
	ec, err := checkVerificationCode(this.Email, this.VerificationCode, orgCLAID)
	if err != nil {
		return ec, err
	}

	return (&this.IndividualSigning).Validate(email)
}

type EmployeeSigningListOption struct {
	CLALanguage string `json:"cla_language"`
}

func (this EmployeeSigningListOption) List(orgRepo *dbmodels.OrgRepo, corpEmail string) ([]dbmodels.IndividualSigningBasicInfo, error) {
	opt := dbmodels.IndividualSigningListOption{
		CLALanguage:      this.CLALanguage,
		CorporationEmail: corpEmail,
	}
	return dbmodels.GetDB().ListIndividualSigning(orgRepo, &opt)
}

type EmployeeSigningUdateInfo struct {
	Enabled bool `json:"enabled"`
}

func (this *EmployeeSigningUdateInfo) Update(orgRepo *dbmodels.OrgRepo, email string) error {
	return dbmodels.GetDB().UpdateIndividualSigning(orgRepo, email, this.Enabled)
}

func DeleteEmployeeSigning(orgRepo *dbmodels.OrgRepo, email string) error {
	return dbmodels.GetDB().DeleteIndividualSigning(orgRepo, email)
}
