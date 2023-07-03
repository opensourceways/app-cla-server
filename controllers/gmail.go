package controllers

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/gmailimpl"
)

const authURLState = "state-token-cla"

type GmailController struct {
	baseController
}

func (this *GmailController) Prepare() {
	this.stopRunIfSignSerivceIsUnabled()

	if !strings.HasSuffix(this.routerPattern(), "auth") {
		this.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Auth
// @Description authorized by org email
// @router /auth [get]
func (this *GmailController) Auth() {
	cfg := &config.AppConfig.APIConfig

	rs := func(errCode string, reason error) {
		this.setCookies(map[string]string{"error_code": errCode, "error_msg": reason.Error()})
		this.redirect(cfg.WebRedirectDirOnFailureForEmail)
	}

	if err := this.GetString("error"); err != "" {
		rs(errAuthFailed, fmt.Errorf("%s, %s", err, this.GetString("error_description")))
		return
	}

	if this.GetString("state") != authURLState {
		rs(errSystemError, fmt.Errorf("unkown state"))
		return
	}

	addr, err := models.AddGmailCredential(
		this.GetString("code"), this.GetString("scope"),
	)
	if err != nil {
		rs(parseModelError(err).errCode, err)

		return
	}

	this.setCookies(map[string]string{"email": addr})
	this.redirect(cfg.WebRedirectDirOnSuccessForEmail)
}

// @Title Get
// @Description get auth code url
// @router /authcodeurl [get]
func (this *GmailController) Get() {
	this.sendSuccessResp(map[string]string{
		"url": gmailimpl.GmailClient().GetOauth2CodeURL(authURLState),
	})
}
