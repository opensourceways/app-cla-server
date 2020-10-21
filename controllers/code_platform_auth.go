package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
	"github.com/opensourceways/app-cla-server/code-platform-auth/platforms"
	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/util"
)

type AuthController struct {
	beego.Controller
}

// @Title Auth
// @Description authorized by gitee/github
// @Param	:platform	path 	string				true		"gitee/github"
// @Param	:purpose	path 	string				true		"purpose: login, sign"
// @Success 200
// @router /:platform/:purpose [get]
func (this *AuthController) Auth() {
	rs := func(statusCode int, errCode string, reason error) {
		sendResponse(&this.Controller, statusCode, errCode, reason, nil, "authorized by gitee/github")
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

	at, sc, ec, err := this.newAccessToken(platform, user, purpose, token)
	if err != nil {
		rs(sc, ec, err)
		return
	}

	this.Ctx.SetCookie("access_token", at, "3600", "/")
	this.Ctx.SetCookie("platform_token", token, "3600", "/")

	http.Redirect(this.Ctx.ResponseWriter, this.Ctx.Request, cp.WebRedirectDir(), http.StatusFound)
}

func (this *AuthController) newAccessToken(platform, user, purpose, platformToken string) (string, int, string, error) {
	permission := ""
	switch purpose {
	case "login":
		permission = PermissionOwnerOfOrg
	case "sign":
		permission = PermissionIndividualSigner
	}

	orgm := map[string]bool{}
	if permission == PermissionOwnerOfOrg {
		pt, err := platforms.NewPlatform(platformToken, "", platform)
		if err != nil {
			return "", 400, util.ErrNotSupportedPlatform, err
		}

		if orgs, err := pt.ListOrg(); err == nil {
			for _, item := range orgs {
				orgm[item] = true
			}
		}
	}

	ac := &accessController{
		Expiry:     util.Expiry(conf.AppConfig.APITokenExpiry),
		Permission: permission,
		Payload: &acForCodePlatformPayload{
			accessControllerBasicPayload: accessControllerBasicPayload{
				User: fmt.Sprintf("%s/%s", platform, user),
			},
			PlatformToken: platformToken,
			Orgs:          orgm,
		},
	}

	token, err := ac.NewToken(conf.AppConfig.APITokenKey)
	if err != nil {
		return "", 500, util.ErrSystemError, err
	}

	return token, 0, "", nil
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
