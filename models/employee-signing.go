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
