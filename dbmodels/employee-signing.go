package dbmodels

type EmployeeSigningInfo struct {
	Email   string `json:"email" required:"true"`
	Name    string `json:"name" required:"true"`
	Enabled bool   `json:"enabled"`

	Info map[string]interface{} `json:"info,omitempty"`
}

type EmployeeSigningListOption struct {
	Platform         string `json:"platform" required:"true"`
	OrgID            string `json:"org_id" required:"true"`
	RepoID           string `json:"repo_id,omitempty"`
	CLALanguage      string `json:"cla_language,omitempty"`
	CorporationEmail string `json:"-"`
}

type EmployeeSigningUpdateInfo struct {
	Enabled bool
}
