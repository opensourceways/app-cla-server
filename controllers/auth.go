package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
	"github.com/opensourceways/app-cla-server/conf"
)

type AuthController struct {
	beego.Controller
}

// @Title Get
// @Description get login info
// @Success 200
// @router /:platform/:purpose [get]
func (this *AuthController) Auth() {
	purpose := this.GetString(":purpose")
	if purpose == "" {
		err := fmt.Errorf("missing purpose")
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	code := this.GetString("code")
	if code == "" {
		err := fmt.Errorf("missing code")
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	platform := this.GetString(":platform")
	if platform == "" {
		err := fmt.Errorf("missing platform")
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	//TODO: gitee don't pass the scope parameter
	scope := this.GetString("scope")

	state := this.GetString("state")
	if state != authURLState {
		err := fmt.Errorf("invalid state")
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

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

	tc := &codePlatformAuth{
		accessController: accessController{
			User:       fmt.Sprintf("%s/%s", platform, user),
			Permission: actionToPermission(purpose),
			Expiry:     conf.AppConfig.APITokenExpiry,
		},
		PlatformToken: token,
	}

	at, err := tc.CreateToken(conf.AppConfig.APITokenKey)
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

	platform := this.GetString(":platform")
	if platform == "" {
		reason = fmt.Errorf("missing parameter platform")
		statusCode = 400
		return
	}

	// purpose: login, sign
	purpose := this.GetString(":purpose")
	if purpose == "" {
		reason = fmt.Errorf("missing parameter purpose")
		statusCode = 400
		return
	}

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
