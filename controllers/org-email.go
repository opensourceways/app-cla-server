package controllers

import (
	"fmt"
	"net/http"

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
	} else {
		this.apiPrepare("")
	}

}

// @Title Auth
// @Description authorized by org email
// @Success 200
// @router /auth/:platform [get]
func (this *EmailController) Auth() {
	rs := func(errCode string, reason error) {
		rspOnAuthFailed(&this.Controller, email.EmailAgent.WebRedirectDir(false), errCode, reason)
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

	opt := models.OrgEmail{
		Token:    token,
		Email:    emailAddr,
		Platform: platform,
	}
	if merr := opt.Create(); merr != nil {
		if !merr.IsErrorOf(models.ErrOrgEmailExists) {
			rs(parseModelError(merr).errCode, merr)
			return
		}
	}

	this.Ctx.SetCookie("email", opt.Email, "3600", "/")
	http.Redirect(
		this.Ctx.ResponseWriter, this.Ctx.Request,
		email.EmailAgent.WebRedirectDir(true), http.StatusFound,
	)
}

// @Title Get
// @Description get auth code url
// @Param	platform		path 	string	true		"The email platform"
// @Success 200 {object}
// @router /authcodeurl/:platform [get]
func (this *EmailController) Get() {
	action := "get auth code url of email"
	platform := this.GetString(":platform")

	e, err := email.EmailAgent.GetEmailClient(platform)
	if err != nil {
		this.sendFailedResponse(400, errUnknownEmailPlatform, err, action)
		return
	}

	this.sendResponse(map[string]string{
		"url": e.GetOauth2CodeURL(authURLState),
	}, 0)
}
