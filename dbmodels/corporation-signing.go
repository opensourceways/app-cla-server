package dbmodels

type CorporationSigningCreateOption struct {
	CLAOrgID        string `json:"-"`
	AdminEmail      string `json:"admin_email" required:"true"`
	AdminName       string `json:"admin_name" required:"true"`
	CorporationName string `json:"corporation_name" required:"true"`
	Enabled         bool   `json:"enabled" required:"true"`

	Info map[string]interface{} `json:"info,omitempty"`
}
