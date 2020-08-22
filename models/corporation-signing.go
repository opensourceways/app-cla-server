package models

import "github.com/zengchen1024/cla-server/dbmodels"

type CorporationSigning struct {
	CLAOrgID        string `json:"cla_org_id"`
	AdminEmail      string `json:"admin_email"`
	AdminName       string `json:"admin_name"`
	CorporationName string `json:"corporation_name"`
	Enabled         bool   `json:"enabled"`

	Info map[string]interface{} `json:"info"`
}

func (this *CorporationSigning) Create() error {
	p := dbmodels.CorporationSigningInfo{
		AdminEmail:      this.AdminEmail,
		AdminName:       this.AdminName,
		CorporationName: this.CorporationName,
		CorporationID:   emailSuffixToKey(this.AdminEmail),
		Enabled:         false,
		Info:            this.Info,
	}
	return dbmodels.GetDB().SignAsCorporation(this.CLAOrgID, p)
}

type CorporationSigningUdateInfo struct {
	CLAOrgID        string `json:"cla_org_id"`
	AdminEmail      string `json:"admin_email"`
	CorporationName string `json:"corporation_name"`
	Enabled         bool   `json:"enabled"`
}

func (this *CorporationSigningUdateInfo) Update() error {
	return dbmodels.GetDB().UpdateCorporationOfOrg(
		this.CLAOrgID, this.AdminEmail, this.CorporationName,
		dbmodels.CorporationSigningUpdateInfo{Enabled: &this.Enabled})
}

type CorporationSigningListOption struct {
	Platform    string `json:"platform"`
	OrgID       string `json:"org_id"`
	RepoID      string `json:"repo_id"`
	CLALanguage string `json:"cla_language"`
}

func (this CorporationSigningListOption) List() ([]CorporationSigning, error) {
	opt := dbmodels.CorporationSigningListOption{
		Platform:    this.Platform,
		OrgID:       this.OrgID,
		RepoID:      this.RepoID,
		CLALanguage: this.CLALanguage,
		ApplyTo:     ApplyToCorporation,
	}
	v, err := dbmodels.GetDB().ListCorporationsOfOrg(opt)
	if err != nil {
		return nil, err
	}

	n := 0
	for _, items := range v {
		n += len(items)
	}

	r := make([]CorporationSigning, 0, n)
	for k, items := range v {
		for _, item := range items {
			r = append(r, CorporationSigning{
				CLAOrgID:        k,
				AdminEmail:      item.AdminEmail,
				AdminName:       item.AdminName,
				CorporationName: item.CorporationName,
				Enabled:         item.Enabled,
			})
		}
	}
	return r, nil
}
