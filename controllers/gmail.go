package controllers

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/gmailimpl"
)

const authURLState = "state-token-cla"

type GmailController struct {
	baseController
}

func (ctl *GmailController) Prepare() {
	if !strings.HasSuffix(ctl.routerPattern(), "auth") {
		ctl.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Callback
// @Description callback of gmail
// @Tags Gmail
// @Accept json
// @router /auth [get]
func (ctl *GmailController) Callback() {
	rs := func(errCode string, reason error) {
		ctl.setCookies(map[string]string{"error_code": errCode, "error_msg": reason.Error()})
		ctl.redirect(config.WebRedirectDirOnFailureForEmail)
	}

	if err := ctl.GetString("error"); err != "" {
		rs(errAuthFailed, fmt.Errorf("%s, %s", err, ctl.GetString("error_description")))
		return
	}

	if ctl.GetString("state") != authURLState {
		rs(errSystemError, fmt.Errorf("unkown state"))
		return
	}

	addr, err := models.AuthorizeGmail(
		ctl.GetString("code"), ctl.GetString("scope"),
	)
	if err != nil {
		rs(parseModelError(err).errCode, err)

		return
	}

	ctl.setCookies(map[string]string{"email": addr})
	ctl.redirect(config.WebRedirectDirOnSuccessForEmail)
}

// @Title AuthCodeURL
// @Description get auth code url
// @Tags Gmail
// @Accept json
// @Success 200 {object} controllers.authCodeURL
// @router /authcodeurl [get]
func (ctl *GmailController) AuthCodeURL() {
	ctl.sendSuccessResp(authCodeURL{
		gmailimpl.GmailClient().GetOauth2CodeURL(authURLState),
	})
}
