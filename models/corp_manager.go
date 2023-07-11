package models

import "fmt"

const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
)

type CorporationManagerLoginInfo struct {
	User     string `json:"user"`
	LinkID   string `json:"link_id"`
	Password string `json:"password"`
}

func (info *CorporationManagerLoginInfo) Validate() IModelError {
	if info.LinkID == "" || info.Password == "" || info.User == "" {
		return newModelError(ErrEmptyPayload, fmt.Errorf("necessary parameters is empty"))
	}

	return nil
}

type CorporationManagerChangePassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type CorpManagerLoginInfo struct {
	Role             string
	Email            string
	UserId           string
	CorpName         string
	SigningId        string
	InitialPWChanged bool
}

type CorporationManagerCreateOption struct {
	ID       string
	Name     string
	Role     string
	Email    string
	Password string
}

type CorporationManagerListResult struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
