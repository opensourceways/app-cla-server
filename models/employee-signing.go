package models

import "github.com/zengchen1024/cla-server/dbmodels"

type EmployeeSigning struct {
	CLAOrgID string                 `json:"cla_org_id"`
	Email    string                 `json:"email"`
	Name     string                 `json:"name"`
	Enabled  bool                   `json:"enabled"`
	Info     map[string]interface{} `json:"info,omitempty"`
}

func (this *EmployeeSigning) Create() error {
	p := dbmodels.EmployeeSigningInfo{
		Email:   this.Email,
		Name:    this.Name,
		Enabled: false,
		Info:    this.Info,
	}
	return dbmodels.GetDB().SignAsEmployee(this.CLAOrgID, p)
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
