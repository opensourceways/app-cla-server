package models

import (
	"encoding/json"
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"golang.org/x/oauth2"
)

type OrgEmail struct {
	Email string `json:"email"`
	// Platform is the email platform, such as gmail
	Platform string        `json:"platform"`
	Token    *oauth2.Token `json:"token"`
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

func GetOrgEmailInfo(email string) (*OrgEmail, IModelError) {
	info, err := dbmodels.GetDB().GetOrgEmailInfo(email)
	if err != nil {
		if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
			return nil, newModelError(ErrOrgEmailNotExists, err)
		}
		return nil, parseDBError(err)
	}

	var token oauth2.Token

	if err := json.Unmarshal(info.Token, &token); err != nil {
		return nil, newModelError(ErrSystemError, fmt.Errorf("Failed to unmarshal oauth2 token: %s", err.Error()))
	}

	return &OrgEmail{
		Email:    email,
		Token:    &token,
		Platform: info.Platform,
	}, nil
}

type OrgEmailAuthorizeCodeModel struct {
	Email         string `json:"email"`
	Platform      string `json:"platform"`
	AuthorizeCode string `json:"authorize_code"`
}

func (this *OrgEmailAuthorizeCodeModel) Create() IModelError {

	opt := dbmodels.OrgEmailCreateInfo{
		Email:         this.Email,
		Platform:      this.Platform,
		AuthorizeCode: this.AuthorizeCode,
	}
	dbErr := dbmodels.GetDB().CreateOrgEmail(opt)
	return parseDBError(dbErr)
}

func GetOrgEmailInfoToSendMail(orgEmail string) (*email.SendInfo, IModelError) {
	info, err := dbmodels.GetDB().GetOrgEmailInfo(orgEmail)
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

	return &email.SendInfo{
		Email:         info.Email,
		Platform:      info.Platform,
		Token:         &token,
		AuthorizeCode: info.AuthorizeCode,
	}, nil
}
