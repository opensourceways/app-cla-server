package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
)

const authURLState = "state-token-cla"

type EmailController struct {
	beego.Controller
}

// @Title Send Email by org
// @Description send email byal org
// @Param	body		body 	emails.EmailMessage	true		"body for email"
// @Failure 403 body is empty
// @router / [post]
func (this *EmailController) Post() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var msg email.EmailMessage
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &msg); err != nil {
		reason = fmt.Errorf("Parse email parameter failed: %s", err.Error())
		statusCode = 400
		return
	}

	cfg := &models.OrgEmail{Email: msg.From}
	if err := cfg.Get(); err != nil {
		reason = fmt.Errorf("Failed to get email cfg: %s", err.Error())
		statusCode = 400
		return
	}

	e, err := email.GetEmailClient(cfg.Platform)
	if err != nil {
		reason = fmt.Errorf("Failtd to get email client: %s", err.Error())
		statusCode = 500
		return
	}

	if err := e.SendEmail(cfg.Token, &msg); err != nil {
		reason = fmt.Errorf("Failed to send email: %s", err.Error())
		statusCode = 500
		return
	}

	body = "send email successfully"
}

// @Title Get
// @Description get login info
// @Success 200
// @router /auth/:platform [get]
func (this *EmailController) Auth() {
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

	platform := this.GetString(":platform")
	if platform == "" {
		err := fmt.Errorf("missing platform")
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	e, err := email.GetEmailClient(platform)
	if err != nil {
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	opt, err := e.GetAuthorizedEmail(code, scope)
	if err != nil {
		sendResponse(&this.Controller, 400, err, nil)
		return
	}
	opt.Platform = platform

	if err = opt.Create(); err != nil {
		sendResponse(&this.Controller, 500, err, nil)
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
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	err := checkApiAccessToken(&this.Controller, []string{PermissionOwnerOfOrg}, &accessController{})
	if err != nil {
		reason = err
		statusCode = 400
		return
	}

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
