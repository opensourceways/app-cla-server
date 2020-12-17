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

func (this *OrgEmail) Create() error {
	b, err := json.Marshal(this.Token)
	if err != nil {
		return fmt.Errorf("Failed to marshal oauth2 token: %s", err.Error())
	}

	opt := dbmodels.OrgEmailCreateInfo{
		Email:    this.Email,
		Platform: this.Platform,
		Token:    b,
	}
	return dbmodels.GetDB().CreateOrgEmail(opt)
}

func (this *OrgEmail) Get() error {
	info, dbErr := dbmodels.GetDB().GetOrgEmailInfo(this.Email)
	if dbErr != nil {
		if dbErr.IsErrorOf(dbmodels.ErrNoDBRecord) {
			return newModelError(ErrOrgEmailNotExist, dbErr.Err)
		}
		return parseDBError(dbErr)
	}

	var token oauth2.Token

	if err := json.Unmarshal(info.Token, &token); err != nil {
		return fmt.Errorf("Failed to unmarshal oauth2 token: %s", err.Error())
	}

	this.Token = &token
	this.Platform = info.Platform

	return nil
}
