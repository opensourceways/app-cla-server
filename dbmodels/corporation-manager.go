package dbmodels

const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
)

type CorporationManagerCreateOption struct {
	Name     string `json:"name"`
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
	Role             string
	Name             string
	Email            string
	InitialPWChanged bool

	Platform string
	OrgID    string
	RepoID   string
}

type CorporationManagerListResult struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
