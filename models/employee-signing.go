package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type EmployeeSigningListOption struct {
	CLALanguage string `json:"cla_language"`
}

func (this EmployeeSigningListOption) List(corpEmail, platform, org, repo string) (map[string][]dbmodels.IndividualSigningBasicInfo, error) {
	opt := dbmodels.IndividualSigningListOption{
		Platform:         platform,
		OrgID:            org,
		RepoID:           repo,
		CLALanguage:      this.CLALanguage,
		CorporationEmail: corpEmail,
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
