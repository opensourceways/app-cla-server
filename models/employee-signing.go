package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type EmployeeSigning struct {
	CLAOrgID string                   `json:"cla_org_id"`
	Email    string                   `json:"email"`
	Name     string                   `json:"name"`
	Enabled  bool                     `json:"enabled"`
	Info     dbmodels.TypeSigningInfo `json:"info,omitempty"`
}

func (this *EmployeeSigning) Create(claOrgID string) error {
	p := dbmodels.EmployeeSigningInfo{
		Email:   this.Email,
		Name:    this.Name,
		Enabled: false,
		Info:    this.Info,
	}
	return dbmodels.GetDB().SignAsEmployee(claOrgID, p)
}

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
	return dbmodels.GetDB().DeleteEmployeeSigning(claOrgID, email)
}
