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

type CorporationManagerChangePassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type CorporationManagerListResult struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
