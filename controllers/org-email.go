package controllers

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
)

const authURLState = "state-token-cla"

type EmailController struct {
	baseController
}

func (this *EmailController) Prepare() {
	if strings.HasSuffix(this.routerPattern(), "authcodeurl/:platform") {
		this.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Auth
// @Description authorized by org email
// @router /auth/:platform/:way [get]
func (this *EmailController) Auth() {
	way := this.GetString(":way")
	rs := func(errCode string, reason error) {
		this.setCookies(map[string]string{"error_code": errCode, "error_msg": reason.Error()}, false)
		this.redirect(email.EmailAgent.WebRedirectDir(false, way))
	}

	if err := this.GetString("error"); err != "" {
		rs(errAuthFailed, fmt.Errorf("%s, %s", err, this.GetString("error_description")))
		return
	}

	platform := this.GetString(":platform")
	emailClient, err := email.EmailAgent.GetEmailClient(platform)
	if err != nil {
		rs(errUnsupportedEmailPlatform, err)
		return
	}

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
		if email.IsReauthType(way) {
			rs(errNoRefreshToken, fmt.Errorf("no refresh token"))
			return
		}
		v, err := models.GetOrgEmailInfo(emailAddr)
		if err != nil {
			rs(errSystemError, err)
			return
		}
		if v == nil {
			rs(errNoRefreshToken, fmt.Errorf("no refresh token"))
			return
		}
	} else {
		opt := models.OrgEmail{
			Token:    token,
			Email:    emailAddr,
			Platform: platform,
		}
		if err := opt.Create(); err != nil {
			rs(parseModelError(err).errCode, err)
			return
		}
	}

	this.setCookies(map[string]string{"email": emailAddr}, false)
	webUrl := email.EmailAgent.WebRedirectDir(true, way)
	this.redirect(webUrl)
}

// @Title Get
// @Description get auth code url
// @Param   platform      path		string	true	"The email platform"
// @Param   way           path		string	false	"the value is init_auth or reauth"
// @router /authcodeurl/:platform/:way [patch]
func (this *EmailController) Switch() {
	action := "get auth code url of email"
	e, err := email.EmailAgent.GetEmailClient(this.GetString(":platform"))
	if err != nil {
		this.sendFailedResponse(400, errUnknownEmailPlatform, err, action)
		return
	}

	way := this.GetString(":way")
	if !email.IsValidAuthType(way) {
		this.sendFailedResponse(400, errNotSupportAuthWay, err, action)
		return
	}

	this.sendSuccessResp(map[string]string{
		"url": e.SwitchAuthType(authURLState, way),
	})
}
