package controllers

import (
	"fmt"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/email"
)

const authURLState = "state-token-cla"

type EmailController struct {
	beego.Controller
}

// @Title Get
// @Description get login info
// @Success 200
// @router /orgemail-gmail [get]
func (this *EmailController) CallBack() {
	code := this.GetString("code")
	if code == "" {
		err := fmt.Errorf("missing code")
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	scope := this.GetString("scope")
	if scope == "" {
		err := fmt.Errorf("missing scope")
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	state := this.GetString("state")
	if state != authURLState {
		err := fmt.Errorf("invalid state")
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	e, err := email.GetEmailClient("gmail")
	if err != nil {
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	opt, err := e.GetAuthorizedEmail(code, scope)
	if err != nil {
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	if err = opt.Create(); err != nil {
		sendResponse(&this.Controller, 500, err, nil)
		return
	}
}

// @Title Get
// @Description get auth code url
// @Param	platform		path 	string	true		"The email platform"
// @Success 200 {object}
// @Failure 403 :platform is empty
// @router /code-url/:platform [get]
func (this *EmailController) Get() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
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
