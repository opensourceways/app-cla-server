package dbmodels

const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
)

type CorporationManagerCreateOption struct {
	Role     string `json:"role"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CorporationManagerCheckInfo struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type CorporationManagerResetPassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type CorporationManagerCheckResult struct {
	Role             string `json:"role"`
	Email            string `json:"email"`
	Platform         string `json:"platform"`
	OrgID            string `json:"org_id"`
	RepoID           string `json:"repo_id"`
	InitialPWChanged bool   `json:"initial_pw_changed"`
}

type CorporationManagerListResult struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
