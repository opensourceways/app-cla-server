package models

import (
	"fmt"

	"golang.org/x/oauth2"
)

type OrgEmail struct {
	Email string `json:"email"`
	// Platform is the email platform, such as gmail
	Platform string        `json:"platform"`
	Token    *oauth2.Token `json:"token"`
	AuthCode string        `json:"auth_code"`
}

type EmailAuthorizationReq struct {
	Email     string `json:"email"`
	Authorize string `json:"authorize"`
}

type EmailAuthorization struct {
	Code string `json:"code"`
	EmailAuthorizationReq
}

func PurposeOfEmailAuthorization(email string) string {
	return fmt.Sprintf("email authorization: %s", email)
}
