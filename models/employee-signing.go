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

func (this *EmployeeSigningUdateInfo) Update(platform, org, repo, email string) error {
	return dbmodels.GetDB().UpdateIndividualSigning(
		platform, org, repo, email, this.Enabled,
	)
}

func DeleteEmployeeSigning(platform, org, repo, email string) error {
	return dbmodels.GetDB().DeleteIndividualSigning(platform, org, repo, email)
}
