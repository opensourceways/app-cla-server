package models

import (
	"encoding/json"
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"golang.org/x/oauth2"
)

type OrgEmail struct {
	Email string `json:"email"`
	// Platform is the email platform, such as gmail
	Platform string        `json:"platform"`
	Token    *oauth2.Token `json:"token"`
	AuthCode string        `json:"auth_code"`
}

func (this *OrgEmail) Create() IModelError {
	b, err := json.Marshal(this.Token)
	if err != nil {
		return newModelError(ErrSystemError, fmt.Errorf("Failed to marshal oauth2 token: %s", err.Error()))
	}

	opt := dbmodels.OrgEmailCreateInfo{
		Email:    this.Email,
		Platform: this.Platform,
		Token:    b,
	}
	dbErr := dbmodels.GetDB().CreateOrgEmail(opt)
	return parseDBError(dbErr)
}

func (this *OrgEmail) CreateUseAuthCode() IModelError {
	opt := dbmodels.OrgEmailCreateInfo{
		Email:    this.Email,
		Platform: this.Platform,
		AuthCode: this.AuthCode,
	}
	dbErr := dbmodels.GetDB().CreateOrgEmail(opt)
	return parseDBError(dbErr)
}

func GetOrgEmailInfo(email string) (*OrgEmail, IModelError) {
	info, err := dbmodels.GetDB().GetOrgEmailInfo(email)
	if err != nil {
		if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
			return nil, newModelError(ErrOrgEmailNotExists, err)
		}
		return nil, parseDBError(err)
	}

	var token oauth2.Token

	if len(info.Token) > 0 {
		if err := json.Unmarshal(info.Token, &token); err != nil {
			return nil, newModelError(ErrSystemError, fmt.Errorf("Failed to unmarshal oauth2 token: %s", err.Error()))
		}
	}

	return &OrgEmail{
		Email:    email,
		Token:    &token,
		Platform: info.Platform,
		AuthCode: info.AuthCode,
	}, nil
}

type EmailAuthorizationReq struct {
	Email     string `json:"email"`
	Authorize string `json:"authorize"`
}

type EmailAuthorization struct {
	Code string `json:"code"`
	EmailAuthorizationReq
}

func (e *EmailAuthorization) Validate() IModelError {
	return validateCodeForSettingOrgEmail(e.Email, e.Code)
}

func PurposeOfEmailAuthorization(email string) string {
	return fmt.Sprintf("email authorization: %s", email)
}

func AddTxmailCredential(opt *EmailAuthorizationReq) IModelError {
	return emailCredentialAdapterInstance.AddTXmailCredential(
		opt.Email, opt.Authorize,
	)
}

func AddGmailCredential(code, scope string) (string, IModelError) {
	return emailCredentialAdapterInstance.AddGmailCredential(code, scope)
}
