package models

import "github.com/zengchen1024/cla-server/dbmodels"

type CorporationManagerCreateOption struct {
	CLAOrgID        string `json:"cla_org_id"`
	AdminEmail      string `json:"admin_email"`
	AdminName       string `json:"admin_name"`
	CorporationName string `json:"corporation_name"`
}

func (this *CorporationManagerCreateOption) Create() error {
	pw := "123456"
	return dbmodels.GetDB().UpdateCorporationOfOrg(
		this.CLAOrgID, this.AdminEmail, this.CorporationName,
		dbmodels.CorporationSigningUpdateInfo{Password: pw})
}

type CorporationManagerAuthentication struct {
	Platform string `json:"platform"`
	OrgID    string `json:"org_id"`
	RepoID   string `json:"repo_id"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func (this CorporationManagerAuthentication) Authenticate() error {
	opt := dbmodels.CorporationManagerCheckInfo{
		Platform: this.Platform,
		OrgID:    this.OrgID,
		RepoID:   this.RepoID,
		User:     this.User,
		Password: this.Password,
	}

	_, err := dbmodels.GetDB().CheckCorporationManagerExist(opt)
	return err
}
