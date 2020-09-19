package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
)

type AuthController struct {
	beego.Controller
}

// @Title Get
// @Description get login info
// @Success 200
// @router /:platform/:purpose [get]
func (this *AuthController) Auth() {
	params := map[string]string{":platform": "", "code": "", ":purpose": "", "state": authURLState}
	if err := checkAndVerifyAPIStringParameter(&this.Controller, params); err != nil {
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	purpose := this.GetString(":purpose")
	code := this.GetString("code")
	platform := this.GetString(":platform")
	//TODO: gitee don't pass the scope parameter
	scope := this.GetString("scope")

	cp, err := platformAuth.GetAuthInstance(platform, purpose)
	if err != nil {
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	token, user, err := cp.Auth(code, scope)
	if err != nil {
		err = fmt.Errorf("Failed to auth: %s", err.Error())
		sendResponse(&this.Controller, 500, err, nil)
		return
	}

	at, err := newAccessTokenAuthorizedByCodePlatform(
		fmt.Sprintf("%s/%s", platform, user),
		actionToPermission(purpose),
		token,
	)
	if err != nil {
		sendResponse(&this.Controller, 500, err, nil)
		return
	}

	this.Ctx.SetCookie("access_token", at, "3600", "/")
	this.Ctx.SetCookie("platform_token", token, "3600", "/")

	http.Redirect(this.Ctx.ResponseWriter, this.Ctx.Request, cp.WebRedirectDir(), http.StatusFound)
}

// @Title Get
// @Description get auth code url
// @Success 200 {object}
// @router /authcodeurl/:platform/:purpose [get]
func (this *AuthController) Get() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	params := []string{":platform", ":purpose"}
	if err := checkAPIStringParameter(&this.Controller, params); err != nil {
		reason = err
		statusCode = 400
		return
	}

	platform := this.GetString(":platform")
	// purpose: login, sign
	purpose := this.GetString(":purpose")
	cp, err := platformAuth.GetAuthInstance(platform, purpose)
	if cp == nil {
		reason = err
		statusCode = 400
		return
	}

	body = map[string]string{
		"url": cp.GetAuthCodeURL(authURLState),
	}
}
