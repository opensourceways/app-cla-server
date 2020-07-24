package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"golang.org/x/oauth2"

	"github.com/zengchen1024/cla-server/controllers/platforms"
)

type LoginController struct {
	beego.Controller
}

// @Title Get
// @Description get login info
// @Success 200
// @router / [get]
func (this *LoginController) Get() {
	code := this.GetString("code")
	platform := this.GetString("platform")

	if code == "" {
		err := fmt.Errorf("missing code")
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	if platform == "" {
		err := fmt.Errorf("missing platform")
		sendResponse(&this.Controller, 400, err, nil)
		return
	}

	token, err := getToken(code, platform)
	if err != nil {
		err = fmt.Errorf("get token failed: %s", err.Error())
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

	redirectUrl := beego.AppConfig.String("client_redirect_url")
	http.Redirect(this.Ctx.ResponseWriter, this.Ctx.Request, redirectUrl, http.StatusFound)
}

func setCookie(w http.ResponseWriter, key string, value string) {
	cookie := http.Cookie{Name: key, Value: value, Path: "/", MaxAge: 86400}
	http.SetCookie(w, &cookie)
}

func getToken(code, platform string) (*oauth2.Token, error) {
	endpoint, err := platforms.GetOauthEndpoint(platform)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s?platform=%s", beego.AppConfig.String("server_redirect_url"), platform)

	cfg := oauth2.Config{
		ClientID:     beego.AppConfig.String(fmt.Sprintf("%s::cla_client_id", platform)),
		ClientSecret: beego.AppConfig.String(fmt.Sprintf("%s::cla_secret_id", platform)),
		Scopes:       []string{"emails", "user_info"},
		Endpoint:     endpoint,
		RedirectURL:  url,
	}

	return cfg.Exchange(context.Background(), code)
}
