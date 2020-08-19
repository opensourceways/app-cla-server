package models

import "github.com/zengchen1024/cla-server/dbmodels"

type CorporationSigning struct {
	CLAOrgID        string `json:"cla_org_id"`
	AdminEmail      string `json:"admin_email"`
	AdminName       string `json:"admin_name"`
	CorporationName string `json:"corporation_name"`
	Info            map[string]interface{}
}

func (this *CorporationSigning) Create() error {
	p := dbmodels.CorporationSigningCreateOption{
		CLAOrgID:        this.CLAOrgID,
		AdminEmail:      this.AdminEmail,
		AdminName:       this.AdminName,
		CorporationName: this.CorporationName,
		Enabled:         false,
		Info:            this.Info,
	}
	return dbmodels.GetDB().SignAsCorporation(p)
}
