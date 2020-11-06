package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
	"github.com/opensourceways/app-cla-server/code-platform-auth/platforms"
	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/util"
)

type AuthController struct {
	beego.Controller
}

// @Title Auth
// @Description authorized by gitee/github
// @Param	:platform	path 	string				true		"gitee/github"
// @Param	:purpose	path 	string				true		"purpose: login, sign"
// @Success 200
// @router /:platform/:purpose [get]
func (this *AuthController) Auth() {
	purpose := this.GetString(":purpose")
	platform := this.GetString(":platform")
	authHelper, ok := platformAuth.Auth[purpose]
	if !ok {
		return
	}

	if this.GetString("state") != authURLState {
		return
	}

	rs := func(errCode string, reason error) {
		this.Ctx.SetCookie("error_code", errCode, "3600", "/")
		this.Ctx.SetCookie("error_msg", reason.Error(), "3600", "/")

		http.Redirect(
			this.Ctx.ResponseWriter, this.Ctx.Request,
			authHelper.WebRedirectDir(false), http.StatusFound,
		)
	}

	if err := this.GetString("error"); err != "" {
		rs(util.ErrAuthFailed, fmt.Errorf("%s, %s", err, this.GetString("error_description")))
		return
	}

	cp, err := authHelper.GetAuthInstance(platform)
	if err != nil {
		rs(util.ErrNotSupportedPlatform, err)
		return
	}

	// gitee don't pass the scope paramter
	token, err := cp.GetToken(this.GetString("code"), this.GetString("scope"))
	if err != nil {
		rs(util.ErrSystemError, err)
		return
	}

	at, ec, err := this.newAccessToken(platform, purpose, token)
	if err != nil {
		rs(ec, err)
		return
	}

	this.Ctx.SetCookie("access_token", at, "3600", "/")
	this.Ctx.SetCookie("platform_token", token, "3600", "/")

	http.Redirect(
		this.Ctx.ResponseWriter, this.Ctx.Request,
		authHelper.WebRedirectDir(true), http.StatusFound,
	)
}

func (this *AuthController) newAccessToken(platform, purpose, platformToken string) (string, string, error) {
	permission := ""
	switch purpose {
	case platformAuth.AuthApplyToLogin:
		permission = PermissionOwnerOfOrg
	case platformAuth.AuthApplyToSign:
		permission = PermissionIndividualSigner
	}

	pt, err := platforms.NewPlatform(platformToken, "", platform)
	if err != nil {
		return "", util.ErrNotSupportedPlatform, err
	}

	orgm := map[string]bool{}
	if permission == PermissionOwnerOfOrg {
		if orgs, err := pt.ListOrg(); err == nil {
			for _, item := range orgs {
				orgm[item] = true
			}
		}
	}

	user, err := pt.GetUser()
	if err != nil {
		return "", util.ErrSystemError, err
	}

	ac := &accessController{
		Expiry:     util.Expiry(conf.AppConfig.APITokenExpiry),
		Permission: permission,
		Payload: &acForCodePlatformPayload{
			User:          user,
			Platform:      platform,
			PlatformToken: platformToken,
			Orgs:          orgm,
		},
	}

	token, err := ac.NewToken(conf.AppConfig.APITokenKey)
	if err != nil {
		return "", util.ErrSystemError, err
	}

	return token, "", nil
}

// @Title Get
// @Description get auth code url
// @Param	:platform	path 	string				true		"gitee/github"
// @Param	:purpose	path 	string				true		"purpose: login, sign"
// @Success 200 {object}
// @Failure util.ErrNotSupportedPlatform
// @router /authcodeurl/:platform/:purpose [get]
func (this *AuthController) Get() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "fetch auth code url of gitee/github")
	}()

	authHelper, ok := platformAuth.Auth[this.GetString(":purpose")]
	if !ok {
		reason = fmt.Errorf("unkonw purpose")
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	cp, err := authHelper.GetAuthInstance(this.GetString(":platform"))
	if err != nil {
		reason = err
		errCode = util.ErrNotSupportedPlatform
		statusCode = 400
		return
	}

	body = map[string]string{
		"url": cp.GetAuthCodeURL(authURLState),
	}
}
