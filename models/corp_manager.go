package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

type CorporationManagerAuthentication struct {
	User     string `json:"user"`
	Password string `json:"password"`
	LinkID   string `json:"link_id"`
}

func (this CorporationManagerAuthentication) Validate() IModelError {
	if this.LinkID == "" || this.Password == "" || this.User == "" {
		return newModelError(ErrEmptyPayload, fmt.Errorf("necessary parameters is empty"))
	}

	return nil
}

type CorporationManagerChangePassword dbmodels.CorporationManagerChangePassword

type CorpManagerLoginInfo struct {
	Role             string
	Email            string
	UserId           string
	CorpName         string
	SigningId        string
	InitialPWChanged bool
}
