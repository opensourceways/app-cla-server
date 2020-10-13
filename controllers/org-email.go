package controllers

import (
	"net/http"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/util"
)

const authURLState = "state-token-cla"

type EmailController struct {
	beego.Controller
}

func (this *EmailController) Prepare() {
	if getRouterPattern(&this.Controller) == "/v1/email/authcodeurl/:platform" {
		apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, &acForCodePlatform{})
	}
}

// @Title Auth
// @Description authorized by org email
// @Success 200
// @router /auth/:platform [get]
func (this *EmailController) Auth() {
	rs := func(statusCode int, errCode string, err error) {
		sendResponse(&this.Controller, statusCode, errCode, err, nil, "authorized by org email")
	}

	params := map[string]string{":platform": "", "code": "", "scope": "", "state": authURLState}
	if err := checkAndVerifyAPIStringParameter(&this.Controller, params); err != nil {
		rs(400, util.ErrInvalidParameter, err)
		return
	}
	code := this.GetString("code")
	scope := this.GetString("scope")
	platform := this.GetString(":platform")

	e, err := email.GetEmailClient(platform)
	if err != nil {
		rs(400, util.ErrInvalidParameter, err)
		return
	}

	opt, err := e.GetAuthorizedEmail(code, scope)
	if err != nil {
		rs(400, util.ErrInvalidParameter, err)
		return
	}
	opt.Platform = platform

	if err = opt.Create(); err != nil {
		sc, ec := convertDBError(err)
		rs(sc, ec, err)
		return
	}

	this.Ctx.SetCookie("email", opt.Email, "3600", "/")

	http.Redirect(this.Ctx.ResponseWriter, this.Ctx.Request, e.WebRedirectDir(), http.StatusFound)
}

// @Title Get
// @Description get auth code url
// @Param	platform		path 	string	true		"The email platform"
// @Success 200 {object}
// @router /authcodeurl/:platform [get]
func (this *EmailController) Get() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "get auth code url of email")
	}()

	platform, err := fetchStringParameter(&this.Controller, ":platform")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	e, err := email.GetEmailClient(platform)
	if err != nil {
		reason = err
		return
	}

	body = map[string]string{
		"url": e.GetOauth2CodeURL(authURLState),
	}
}
