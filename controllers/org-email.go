package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

const authURLState = "state-token-cla"

type EmailController struct {
	baseController
}

func (this *EmailController) Prepare() {
	if this.routerPattern() == "/v1/email/authcodeurl/:platform" {
		this.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Auth
// @Description authorized by org email
// @Success 200
// @router /auth/:platform [get]
func (this *EmailController) Auth() {
	rs := func(errCode string, reason error) {
		this.setCookies(map[string]string{"error_code": errCode, "error_msg": reason.Error()})
		this.redirect(email.EmailAgent.WebRedirectDir(false))
	}

	if err := this.GetString("error"); err != "" {
		rs(util.ErrAuthFailed, fmt.Errorf("%s, %s", err, this.GetString("error_description")))
		return
	}

	platform := this.GetString(":platform")
	emailClient, err := email.EmailAgent.GetEmailClient(platform)
	if err != nil {
		rs(util.ErrNotSupportedPlatform, err)
		return
	}

	params := map[string]string{"code": "", "scope": "", "state": authURLState}
	if err := checkAndVerifyAPIStringParameter(&this.Controller, params); err != nil {
		rs(util.ErrInvalidParameter, err)
		return
	}

	token, err := emailClient.GetToken(this.GetString("code"), this.GetString("scope"))
	if err != nil {
		rs(util.ErrSystemError, err)
		return
	}

	emailAddr, err := emailClient.GetAuthorizedEmail(token)
	if err != nil {
		rs(util.ErrSystemError, err)
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
		if err = opt.Create(); err != nil {
			rs(util.ErrSystemError, err)
			return
		}
	}

	this.setCookies(map[string]string{"email": emailAddr})
	this.redirect(email.EmailAgent.WebRedirectDir(true))
}

// @Title Get
// @Description get auth code url
// @Param	platform		path 	string	true		"The email platform"
// @Success 200 {object}
// @router /authcodeurl/:platform [get]
func (this *EmailController) Get() {
	sendResp := this.newFuncForSendingFailedResp("get auth code url of email")

	e, err := email.EmailAgent.GetEmailClient(this.GetString(":platform"))
	if err != nil {
		sendResp(newFailedApiResult(400, errUnknownEmailPlatform, err))
		return
	}

	this.sendSuccessResp(map[string]string{
		"url": e.GetOauth2CodeURL(authURLState),
	})
}
