package dbmodels

type IndividualSigningInfo struct {
	Email   string          `json:"email" required:"true"`
	Name    string          `json:"name" required:"true"`
	Enabled bool            `json:"enabled"`
	Date    string          `json:"date" required:"true"`
	Info    TypeSigningInfo `json:"info,omitempty"`
}

type IndividualSigningCheckInfo struct {
	Platform string `json:"platform" required:"true"`
	OrgID    string `json:"org_id" required:"true"`
	RepoID   string `json:"-" required:"true"`
	Email    string `json:"-" required:"true"`
}
