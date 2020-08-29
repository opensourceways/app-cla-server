package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/controllers/platforms"
	"github.com/zengchen1024/cla-server/oauth2"
)

type AuthController struct {
	beego.Controller
}

// @Title Get
// @Description get login info
// @Success 200
// @router /:purpose [get]
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

	platform := this.GetString("platform")
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

	cp, err := oauth2.GetOauth2Instance(platform, purpose)
	if err != nil {
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	token, err := cp.GetToken(code, scope)
	if err != nil {
		err = fmt.Errorf("Get token failed: %s", err.Error())
		sendResponse(&this.Controller, 500, err, nil)
		return
	}

	p, err := platforms.NewPlatform(token.AccessToken, "", platform)
	if err != nil {
		sendResponse(&this.Controller, 500, err, nil)
		return
	}

	user, err := p.GetUser()
	if err != nil {
		err = fmt.Errorf("get %s user failed: %s", platform, err.Error())
		sendResponse(&this.Controller, 500, err, nil)
		return
	}

	setCookie(this.Ctx.ResponseWriter, "access_token", token.AccessToken)
	setCookie(this.Ctx.ResponseWriter, "user", user)

	http.Redirect(this.Ctx.ResponseWriter, this.Ctx.Request, cp.WebRedirectDir(), http.StatusFound)
}

// @Title Get
// @Description get auth code url
// @Success 200 {object}
// @router /authcodeurl [get]
func (this *AuthController) Get() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	platform := this.GetString("platform")
	if platform == "" {
		reason = fmt.Errorf("missing parameter platform")
		statusCode = 400
		return
	}

	// purpose: login, sign
	purpose := this.GetString("purpose")
	if purpose == "" {
		reason = fmt.Errorf("missing parameter purpose")
		statusCode = 400
		return
	}

	cp, err := oauth2.GetOauth2Instance(platform, purpose)
	if cp == nil {
		reason = err
		statusCode = 400
		return
	}

	body = map[string]string{
		"url": cp.GetOauth2CodeURL(authURLState),
	}
}
