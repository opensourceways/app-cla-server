package models

import (
	"encoding/json"
	"fmt"

	"github.com/zengchen1024/cla-server/dbmodels"
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
	info, err := dbmodels.GetDB().GetOrgEmailInfo(this.Email)
	if err != nil {
		return err
	}

	this.Platform = info.Platform

	var token oauth2.Token

	err = json.Unmarshal(info.Token, &token)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal oauth2 token: %s", err.Error())
	}

	this.Token = &token
	return nil
}
