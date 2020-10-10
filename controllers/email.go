package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/email"
)

const authURLState = "state-token-cla"

type EmailController struct {
	beego.Controller
}

func (this *EmailController) Prepare() {
	if getRouterPattern(&this.Controller) == "/v1/email/authcodeurl/:platform" {
		apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, nil)
	}
}

// @Title Get
// @Description get login info
// @Success 200
// @router /auth/:platform [get]
func (this *EmailController) Auth() {
	params := map[string]string{":platform": "", "code": "", "scope": "", "state": authURLState}
	if err := checkAndVerifyAPIStringParameter(&this.Controller, params); err != nil {
		sendResponse1(&this.Controller, 400, err, nil)
		return
	}
	code := this.GetString("code")
	scope := this.GetString("scope")
	platform := this.GetString(":platform")

	e, err := email.GetEmailClient(platform)
	if err != nil {
		sendResponse1(&this.Controller, 400, err, nil)
		return
	}

	opt, err := e.GetAuthorizedEmail(code, scope)
	if err != nil {
		sendResponse1(&this.Controller, 400, err, nil)
		return
	}
	opt.Platform = platform

	if err = opt.Create(); err != nil {
		sendResponse1(&this.Controller, 500, err, nil)
		return
	}

	this.Ctx.SetCookie("email", opt.Email, "3600", "/")

	http.Redirect(this.Ctx.ResponseWriter, this.Ctx.Request, e.WebRedirectDir(), http.StatusFound)
}

// @Title Get
// @Description get auth code url
// @Param	platform		path 	string	true		"The email platform"
// @Success 200 {object}
// @Failure 403 :platform is empty
// @router /authcodeurl/:platform [get]
func (this *EmailController) Get() {
	var statusCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	platform := this.GetString(":platform")
	if platform == "" {
		reason = fmt.Errorf("missing email platform")
		statusCode = 400
		return
	}

	e, err := email.GetEmailClient(platform)
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = map[string]string{
		"url": e.GetOauth2CodeURL(authURLState),
	}
}
