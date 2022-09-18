package controllers

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/worker"

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

func (this *EmailController) Code() {
	action := "send Email authorization verification code"
	platform := this.GetString(":platform")

	var info models.EmailAuthorization
	if fr := this.fetchInputPayloadFromFormData(&info); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	code, ierr := models.CreateVerificationCode(info.Email, models.PurposeOfEmailAuthorization(info.Email), config.AppConfig.VerificationCodeExpiry)
	if ierr != nil {
		this.sendModelErrorAsResp(ierr, action)
		return
	}
	e := email.EmailVerification{
		Code: code,
	}
	msg, err := e.GenEmailMsg()
	if err != nil {
		beego.Error(err)
		return
	}
	msg.From = info.Email
	msg.To = []string{info.Email}
	msg.Subject = "CLA Email authorization verification code"

	worker.GetEmailWorker().SendSimpleMessage(info.Email, platform, info.Authorize, msg)
}

func (this *EmailController) Authorize() {
	action := "Email authorization verification"

	platform := this.GetString(":platform")

	var info models.EmailAuthorization
	if fr := this.fetchInputPayloadFromFormData(&info); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	info.Purpose = models.PurposeOfEmailAuthorization(info.Email)

	if err := (&info).Validate(); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}
	_, err := email.EmailAgent.GetEmailClient(platform)
	if err != nil {
		this.sendFailedResponse(500, errUnknownEmailPlatform, err, action)
		return
	}
	opt := models.OrgEmail{
		Token:     nil,
		Email:     info.Email,
		Platform:  platform,
		Authorize: info.Authorize,
	}

	if ierr := opt.Create(); err != nil {
		this.sendModelErrorAsResp(ierr, action)
		return
	}
	this.setCookies(map[string]string{"email": info.Email})
	this.sendSuccessResp("Email Authorization Success")
	return
}
