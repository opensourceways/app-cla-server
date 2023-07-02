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

	emailClient := gmailimpl.GmailClient()

	if this.GetString("state") != authURLState {
		rs(errSystemError, fmt.Errorf("unkown state"))
		return
	}

	token, err := emailClient.GetToken(this.GetString("code"), this.GetString("scope"))
	if err != nil {
		rs(errSystemError, err)
		return
	}

	emailAddr, err := emailClient.GetAuthorizedEmail(token)
	if err != nil {
		rs(errSystemError, err)
		return
	}

	if token.RefreshToken == "" {
		if _, err := models.GetOrgEmailInfo(emailAddr); err != nil {
			rs(errNoRefreshToken, fmt.Errorf("no refresh token"))
			return
		}
	} else {
		opt := models.OrgEmail{
			Token:    token,
			Email:    emailAddr,
			Platform: gmailimpl.Platform(),
		}
		if err := opt.Create(); err != nil {
			rs(parseModelError(err).errCode, err)
			return
		}
	}

	this.setCookies(map[string]string{"email": emailAddr})
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
