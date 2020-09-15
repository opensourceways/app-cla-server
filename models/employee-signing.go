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

type EmployeeSigningListOption struct {
	Platform         string `json:"platform"`
	OrgID            string `json:"org_id"`
	RepoID           string `json:"repo_id"`
	CLALanguage      string `json:"cla_language"`
	CorporationEmail string `json:"corporation_email"`
}

func (this EmployeeSigningListOption) List() ([]EmployeeSigning, error) {
	opt := dbmodels.EmployeeSigningListOption{
		Platform:         this.Platform,
		OrgID:            this.OrgID,
		RepoID:           this.RepoID,
		CLALanguage:      this.CLALanguage,
		CorporationEmail: this.CorporationEmail,
	}
	v, err := dbmodels.GetDB().ListEmployeeSigning(opt)
	if err != nil {
		return nil, err
	}

	n := 0
	for _, items := range v {
		n += len(items)
	}

	r := make([]EmployeeSigning, 0, n)
	for k, items := range v {
		for _, item := range items {
			r = append(r, EmployeeSigning{
				CLAOrgID: k,
				Email:    item.Email,
				Name:     item.Name,
				Enabled:  item.Enabled,
			})
		}
	}
	return r, nil
}

type EmployeeSigningUdateInfo struct {
	CLAOrgID string `json:"cla_org_id"`
	Email    string `json:"email"`
	Enabled  bool   `json:"enabled"`
}

func (this *EmployeeSigningUdateInfo) Update() error {
	return dbmodels.GetDB().UpdateEmployeeSigning(
		this.CLAOrgID, this.Email,
		dbmodels.EmployeeSigningUpdateInfo{Enabled: this.Enabled},
	)
}

func DeleteEmployeeSigning(claOrgID, email string) error {
	return dbmodels.GetDB().DeleteEmployeeSigning(claOrgID, email)
}
