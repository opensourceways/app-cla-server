package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
	"github.com/opensourceways/app-cla-server/util"
)

type AuthController struct {
	beego.Controller
}

// @Title Get
// @Description get login info
// @Param	:platform	path 	string				true		"gitee/github"
// @Param	:purpose	path 	string				true		"purpose: login, sign"
// @Success 200
// @router /:platform/:purpose [get]
func (this *AuthController) Auth() {
	rs := func(statusCode int, errCode string, reason error) {
		sendResponse(&this.Controller, statusCode, errCode, reason, nil, "authorize by gitee/github")
	}

	params := map[string]string{":platform": "", "code": "", ":purpose": "", "state": authURLState}
	if err := checkAndVerifyAPIStringParameter(&this.Controller, params); err != nil {
		rs(400, util.ErrInvalidParameter, err)
		return
	}

	purpose := this.GetString(":purpose")
	code := this.GetString("code")
	platform := this.GetString(":platform")
	//TODO: gitee don't pass the scope parameter
	scope := this.GetString("scope")

	cp, err := platformAuth.GetAuthInstance(platform, purpose)
	if err != nil {
		rs(400, util.ErrNotSupportedPlatform, err)
		return
	}

	token, user, err := cp.Auth(code, scope)
	if err != nil {
		rs(500, util.ErrSystemError, err)
		return
	}

	at, err := newAccessTokenAuthorizedByCodePlatform(
		fmt.Sprintf("%s/%s", platform, user),
		actionToPermission(purpose),
		token,
	)
	if err != nil {
		rs(500, util.ErrSystemError, err)
		return
	}

	this.Ctx.SetCookie("access_token", at, "3600", "/")
	this.Ctx.SetCookie("platform_token", token, "3600", "/")

	http.Redirect(this.Ctx.ResponseWriter, this.Ctx.Request, cp.WebRedirectDir(), http.StatusFound)
}

// @Title Get
// @Description get auth code url
// @Param	:platform	path 	string				true		"gitee/github"
// @Param	:purpose	path 	string				true		"purpose: login, sign"
// @Success 200 {object}
// @Failure util.ErrNotSupportedPlatform
// @router /authcodeurl/:platform/:purpose [get]
func (this *AuthController) Get() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "fetch auth code url of gitee/github")
	}()

	params := []string{":platform", ":purpose"}
	if err := checkAPIStringParameter(&this.Controller, params); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	purpose := this.GetString(":purpose")
	bingo := false
	for _, v := range []string{"login", "sign"} {
		if v == purpose {
			bingo = true
			break
		}
	}
	if !bingo {
		reason = fmt.Errorf("unkonw purpose")
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	cp, err := platformAuth.GetAuthInstance(this.GetString(":platform"), purpose)
	if cp == nil {
		reason = err
		errCode = util.ErrNotSupportedPlatform
		statusCode = 400
		return
	}

	body = map[string]string{
		"url": cp.GetAuthCodeURL(authURLState),
	}
}
