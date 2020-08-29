package dbmodels

type CLAOrg struct {
	ID          string `json:"id,omitempty"`
	Platform    string `json:"platform" required:"true"`
	OrgID       string `json:"org_id" required:"true"`
	RepoID      string `json:"repo_id" required:"true"`
	CLAID       string `json:"cla_id" required:"true"`
	CLALanguage string `json:"cla_language" required:"true"`
	ApplyTo     string `json:"apply_to" required:"true"`
	OrgEmail    string `json:"org_email" required:"true"`
	Enabled     bool   `json:"enabled"`
	Submitter   string `json:"submitter" required:"true"`
}

type CLAOrgListOption struct {
	Platform string `json:"platform" required:"true"`
	OrgID    string `json:"org_id,omitempty"`
	RepoID   string `json:"-"`
	ApplyTo  string `json:"apply_to,omitempty"`
}
