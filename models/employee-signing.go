package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type EmployeeSigningListOption struct {
	RepoID      string `json:"repo_id"`
	CLALanguage string `json:"cla_language"`
}

func (this EmployeeSigningListOption) List(email, platform, org string) (map[string][]dbmodels.IndividualSigningBasicInfo, error) {
	opt := dbmodels.IndividualSigningListOption{
		Platform:         platform,
		OrgID:            org,
		RepoID:           this.RepoID,
		CLALanguage:      this.CLALanguage,
		CorporationEmail: email,
	}
	return dbmodels.GetDB().ListIndividualSigning(opt)
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
