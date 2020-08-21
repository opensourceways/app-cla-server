package dbmodels

type CorporationManagerCreateOption struct {
	Role          string `json:"role" required:"true"`
	Email         string `json:"email" required:"true"`
	Password      string `json:"password" required:"true"`
	CorporationID string `json:"corporation_id" required:"true"`
}

type CorporationManagerCheckInfo struct {
	Platform string `json:"platform" required:"true"`
	OrgID    string `json:"org_id" required:"true"`
	RepoID   string `json:"repo_id" required:"true"`
	Password string `json:"-"`
	User     string `json:"-"`
}

type CorporationManagerResetPassword struct {
	CorporationManagerCheckInfo
	NewPassword string `json:"-"`
}

type CorporationManagerCheckResult struct {
	CLAOrgID      string `json:"-"`
	Role          string `json:"-"`
	CorporationID string `json:"-"`
}
