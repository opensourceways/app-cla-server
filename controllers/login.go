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
// @Description get all Users
// @Success 200 {object} models.User
// @router / [get]
func (u *LoginController) Get() {
	code := u.GetString("code")
	platform := u.GetString("platform")
	beego.Info("code: ", code, "platform: ", platform)
	if code == "" {
		u.Data["json"] = "received an empty code"
		u.ServeJSON()
		return
	}

	token, err := getToken(code, platform)
	if err != nil {
		beego.Info("get token:", err.Error())

		u.Data["json"] = fmt.Sprintf("get token failed: %s", err.Error())
		u.ServeJSON()
		return
	}

	user, err := getUser(platform, token.AccessToken)
	if err != nil {
		u.Data["json"] = fmt.Sprintf("get user failed: %s", err.Error())
		u.ServeJSON()
		return
	}

	setCookie(u.Ctx.ResponseWriter, "access_token", token.AccessToken)
	setCookie(u.Ctx.ResponseWriter, "refresh_token", token.RefreshToken)
	setCookie(u.Ctx.ResponseWriter, "user", user)

	redirectUrl := beego.AppConfig.String("client_redirect_url")
	http.Redirect(u.Ctx.ResponseWriter, u.Ctx.Request, redirectUrl, http.StatusFound)
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
