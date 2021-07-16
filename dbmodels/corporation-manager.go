package dbmodels

const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
)

type CorporationManagerCreateOption struct {
	ID       string
	Name     string
	Role     string
	Email    string
	Password string
}

type CorporationManagerCheckInfo struct {
	ID          string
	Email       string
	EmailSuffix string
}

type CorporationManagerResetPassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type CorporationManagerCheckResult struct {
	Corp             string
	Role             string
	Name             string
	Email            string
	Password         string
	InitialPWChanged bool

	OrgInfo
}

type CorporationManagerListResult struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
