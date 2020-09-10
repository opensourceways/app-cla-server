package dbmodels

type TypeSigningInfo map[string]string

type CorporationSigningDetails struct {
	CorporationSigningInfo
	AdministratorEnabled bool
}

type CorporationSigningInfo struct {
	AdminEmail      string          `json:"admin_email" required:"true"`
	AdminName       string          `json:"admin_name" required:"true"`
	CorporationName string          `json:"corporation_name" required:"true"`
	CorporationID   string          `json:"corporation_id" required:"true"`
	Enabled         bool            `json:"enabled"`
	Info            TypeSigningInfo `json:"info,omitempty"`
}

type CorporationSigningListOption struct {
	Platform    string `json:"platform" required:"true"`
	OrgID       string `json:"org_id" required:"true"`
	RepoID      string `json:"repo_id"`
	CLALanguage string `json:"cla_language,omitempty"`
}

type CorporationSigningUpdateInfo struct {
	Enabled *bool `json:"enabled,omitempty"`
}
