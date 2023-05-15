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
	Password    string
	LinkID      string
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
	SigningId        string
	InitialPWChanged bool

	OrgInfo
}

type CorporationManagerListResult struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type CorporationDetail struct {
	EmailDomains []string
	Admin        CorporationManagerListResult
	Managers     []CorporationManagerListResult
}

func (d *CorporationDetail) HasDomain(v string) bool {
	for _, item := range d.EmailDomains {
		if item == v {
			return true
		}
	}

	return false
}

func (d *CorporationDetail) IsNotFound() bool {
	return len(d.EmailDomains) == 0
}

func (d *CorporationDetail) AdminEmail() string {
	return d.Admin.Email
}

func (d *CorporationDetail) HasAdmin() bool {
	return d.Admin.Email != ""
}
