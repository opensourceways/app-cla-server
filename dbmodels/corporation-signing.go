package dbmodels

type TypeSigningInfo map[string]string

type CorporationSigningDetails struct {
	CorporationSigningInfo

	AdministratorEnabled bool `json:"administrator_enabled"`
}

type CorporationSigningInfo struct {
	AdminEmail      string          `json:"admin_email"`
	AdminName       string          `json:"admin_name"`
	CorporationName string          `json:"corporation_name"`
	Enabled         bool            `json:"enabled"`
	Date            string          `json:"date"`
	Info            TypeSigningInfo `json:"info"`
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
