package dbmodels

const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
)

type CorporationManagerCreateOption struct {
	Role          string `json:"role" required:"true"`
	Email         string `json:"email" required:"true"`
	Password      string `json:"password" required:"true"`
	CorporationID string `json:"corporation_id" required:"true"`
}

type CorporationManagerCheckInfo struct {
	Password string
	User     string
}

type CorporationManagerResetPassword struct {
	Email       string
	OldPassword string
	NewPassword string
}

type CorporationManagerCheckResult struct {
	CLAOrgID string `json:"cla_org_id"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	Platform string `json:"platform"`
	OrgID    string `json:"org_id"`
	RepoID   string `json:"repo_id"`
}

type CorporationManagerListOption struct {
	Role          string `json:"role"`
	CorporationID string `json:"corporation_id"`
}

type CorporationManagerListResult struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
