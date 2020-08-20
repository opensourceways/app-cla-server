package dbmodels

type CorporationManagerCheckInfo struct {
	Platform string `json:"platform" required:"true"`
	OrgID    string `json:"org_id" required:"true"`
	RepoID   string `json:"repo_id" required:"true"`
	Password string `json:"-"`
	User     string `json:"-"`
}

type CorporationManagerCheckResult struct {
	CLAOrgID      string `json:"-"`
	Role          string `json:"-"`
	CorporationID string `json:"-"`
}
