package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type EmployeeSigningListOption dbmodels.IndividualSigningListOption

func (this EmployeeSigningListOption) List() (map[string][]dbmodels.IndividualSigningBasicInfo, error) {
	return dbmodels.GetDB().ListIndividualSigning(dbmodels.IndividualSigningListOption(this))
}

type EmployeeSigningUdateInfo struct {
	Enabled bool `json:"enabled"`
}

func (this *EmployeeSigningUdateInfo) Update(claOrgID, email string) error {
	return dbmodels.GetDB().UpdateIndividualSigning(
		claOrgID, email, this.Enabled,
	)
}

func DeleteEmployeeSigning(claOrgID, email string) error {
	return dbmodels.GetDB().DeleteIndividualSigning(claOrgID, email)
}
