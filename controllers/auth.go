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
// @router /login [get]
func (this *AuthController) Login() {
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

	cp := oauth2.GetOauth2Instance(platform)
	if cp == nil {
		err := fmt.Errorf("invalide platform")
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	token, err := cp.GetToken(code, scope, "login")
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
	setCookie(this.Ctx.ResponseWriter, "refresh_token", token.RefreshToken)
	setCookie(this.Ctx.ResponseWriter, "user", user)

	redirectUrl := beego.AppConfig.String("web_redirect_dir_login")
	http.Redirect(this.Ctx.ResponseWriter, this.Ctx.Request, redirectUrl, http.StatusFound)
}

// @Title Get
// @Description get auth code url
// @Param	platform		path 	string	true		"The email platform"
// @Success 200 {object}
// @Failure 403 :platform is empty
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

	// target: login, individual signing
	target := this.GetString("target")
	if target == "" {
		reason = fmt.Errorf("missing parameter target")
		statusCode = 400
		return
	}

	cp := oauth2.GetOauth2Instance(platform)
	if cp == nil {
		reason = fmt.Errorf("invalide platform")
		statusCode = 400
		return
	}

	url, err := cp.GetOauth2CodeURL(authURLState, target)
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = map[string]string{
		"url": url,
	}
	beego.Info(body)
}
