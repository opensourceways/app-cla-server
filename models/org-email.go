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

func GetOrgEmailOfLink(linkID string) (*OrgEmail, IModelError) {
	info, err := dbmodels.GetDB().GetOrgEmailOfLink(linkID)
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
		Email:    info.Email,
		Token:    &token,
		Platform: info.Platform,
	}, nil
}

func HasOrgEmail(email string) (bool, IModelError) {
	_, err := dbmodels.GetDB().GetOrgEmailInfo(email)
	if err == nil {
		return true, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return false, nil
	}
	return false, parseDBError(err)
}
