package models

import (
	"encoding/json"
	"fmt"

	"github.com/zengchen1024/cla-server/dbmodels"
	"golang.org/x/oauth2"
)

type OrgEmail struct {
	Email string        `json:"email"`
	Token *oauth2.Token `json:"token"`
}

func (this *OrgEmail) Create() error {
	b, err := json.Marshal(this.Token)
	if err != nil {
		return fmt.Errorf("marshal oauth2 token failed: %s", err.Error())
	}

	opt := dbmodels.OrgEmailCreateInfo{
		Email: this.Email,
		Token: b,
	}
	return dbmodels.GetDB().CreateOrgEmail(opt)
}
