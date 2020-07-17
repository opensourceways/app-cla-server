package controllers

import (
	"context"
	"fmt"
	"net/http"

	"gitee.com/openeuler/go-gitee/gitee"
	"github.com/astaxie/beego"
	"golang.org/x/oauth2"
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
		sendResponse(&this.Controller, 400, err)
		return
	}

	if platform == "" {
		err := fmt.Errorf("missing platform")
		sendResponse(&this.Controller, 400, err)
		return
	}

	token, err := getToken(code, platform)
	if err != nil {
		err = fmt.Errorf("get token failed: %s", err.Error())
		sendResponse(&this.Controller, 500, err)
		return
	}

	user, err := getUser(platform, token.AccessToken)
	if err != nil {
		err = fmt.Errorf("get %s user failed: %s", platform, err.Error())
		sendResponse(&this.Controller, 500, err)
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

func getUser(platform, ak string) (string, error) {
	switch platform {
	case "gitee":
		return getGiteeUser(ak)
	}
	return "", nil
}

func getGiteeUser(ak string) (string, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ak})
	conf := gitee.NewConfiguration()
	conf.HTTPClient = oauth2.NewClient(ctx, ts)
	cli := gitee.NewAPIClient(conf)

	u, _, err := cli.UsersApi.GetV5User(ctx, nil)
	if err != nil {
		return "", err
	}
	return u.Login, err
}

func getToken(code, platform string) (*oauth2.Token, error) {
	url := fmt.Sprintf("%s?platform=%s", beego.AppConfig.String("server_redirect_url"), platform)

	cfg := oauth2.Config{
		ClientID:     beego.AppConfig.String(fmt.Sprintf("%s::cla_client_id", platform)),
		ClientSecret: beego.AppConfig.String(fmt.Sprintf("%s::cla_secret_id", platform)),
		Scopes:       []string{"emails", "user_info"},
		Endpoint:     oauthEndpoint(platform),
		RedirectURL:  url,
	}

	return cfg.Exchange(context.Background(), code)
}

func oauthEndpoint(platform string) oauth2.Endpoint {
	switch platform {
	case "gitee":
		return oauth2.Endpoint{
			AuthURL:  "https://gitee.com/oauth/authorize",
			TokenURL: "https://gitee.com/oauth/token",
		}
	}

	return oauth2.Endpoint{}
}
