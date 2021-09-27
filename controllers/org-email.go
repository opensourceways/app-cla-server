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
	this.stopRunIfSignSerivceIsUnabled()

	if strings.HasSuffix(this.routerPattern(), "authcodeurl/:platform") {
		this.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Auth
// @Description authorized by org email
// @router /auth/:platform [get]
func (this *EmailController) Auth() {
	rs := func(errCode string, reason error) {
		this.setCookies(map[string]string{"error_code": errCode, "error_msg": reason.Error()})
		this.redirect(email.EmailAgent.WebRedirectDir(false))
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
		if _, err := models.GetOrgEmailInfo(emailAddr); err != nil {
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

	this.setCookies(map[string]string{"email": emailAddr})
	this.redirect(email.EmailAgent.WebRedirectDir(true))
}

// @Title Get
// @Description get auth code url
// @Param	platform		path 	string	true		"The email platform"
// @router /authcodeurl/:platform [get]
func (this *EmailController) Get() {
	e, err := email.EmailAgent.GetEmailClient(this.GetString(":platform"))
	if err != nil {
		this.sendFailedResponse(400, errUnknownEmailPlatform, err, "get auth code url of email")
		return
	}

	this.sendSuccessResp(map[string]string{
		"url": e.GetOauth2CodeURL(authURLState),
	})
}
