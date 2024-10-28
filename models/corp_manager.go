package models

import "fmt"

const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
)

type CorporationManagerLoginInfo struct {
	User     string `json:"user"`
	LinkID   string `json:"link_id"`
	Password []byte `json:"password"`
}

func (info *CorporationManagerLoginInfo) Validate() IModelError {
	if info.LinkID == "" || len(info.Password) == 0 || info.User == "" {
		return newModelError(ErrEmptyPayload, fmt.Errorf("necessary parameters is empty"))
	}

	return nil
}

type CorporationManagerChangePassword struct {
	OldPassword []byte `json:"old_password"`
	NewPassword []byte `json:"new_password"`
}

type CorpManagerUserInfo struct {
	Role             string `json:"role"`
	Account          string `json:"account"`
	InitialPWChanged bool   `json:"initial_pw_changed"`
}

type CorpManagerLoginInfo struct {
	Role             string
	Email            string
	UserId           string
	CorpName         string
	SigningId        string
	InitialPWChanged bool
	RetryNum         int
}

type CorporationManagerCreateOption struct {
	ID       string
	Name     string
	Role     string
	Email    string
	Password []byte
}

type CorporationManagerListResult struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
