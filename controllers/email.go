package controllers

import (
	"context"
	"fmt"

	"github.com/astaxie/beego"
	"golang.org/x/oauth2"
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

	path := beego.AppConfig.String("gmail::credentials")

	opt, err := gmailInfo{}.GenOrgEmail(code, path, scope)
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

	cfg, err := this.getOauth2Config(platform)
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = map[string]string{
		"url": getAuthCodeURL(cfg),
	}
}

func (this *EmailController) getOauth2Config(platform string) (*oauth2.Config, error) {
	if platform != "gmail" {
		return nil, fmt.Errorf("it only supports gmail platform currently")
	}

	path := beego.AppConfig.String("gmail::credentials")
	return gmailInfo{}.GetOauth2Config(path)
}

func getAuthCodeURL(cfg *oauth2.Config) string {
	return cfg.AuthCodeURL(authURLState, oauth2.AccessTypeOffline)
}

func fetchOauth2Token(cfg *oauth2.Config, code string) (*oauth2.Token, error) {
	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token: %v", err)
	}
	return token, nil
}
