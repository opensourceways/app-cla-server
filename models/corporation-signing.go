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
	p := dbmodels.CorporationSigningInfo{
		AdminEmail:      this.AdminEmail,
		AdminName:       this.AdminName,
		CorporationName: this.CorporationName,
		Enabled:         false,
		Info:            this.Info,
	}
	return dbmodels.GetDB().SignAsCorporation(this.CLAOrgID, p)
}

type CorporationSigningListOption struct {
	Platform    string `json:"platform"`
	OrgID       string `json:"org_id"`
	RepoID      string `json:"repo_id"`
	CLALanguage string `json:"cla_language"`
}

func (this CorporationSigningListOption) List() ([]dbmodels.CorporationSigningInfo, error) {
	opt := dbmodels.CorporationSigningListOption{
		Platform:    this.Platform,
		OrgID:       this.OrgID,
		RepoID:      this.RepoID,
		CLALanguage: this.CLALanguage,
		ApplyTo:     ApplyToCorporation,
	}
	return dbmodels.GetDB().ListCorporationsOfOrg(opt)
}
