package controllers

import (
	"errors"
	"fmt"
	"strings"

	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
	"github.com/opensourceways/app-cla-server/code-platform-auth/platforms"
	"github.com/opensourceways/app-cla-server/models"
)

type AuthController struct {
	baseController
}

func (ctl *AuthController) Prepare() {
	if strings.HasSuffix(ctl.routerPattern(), "/authcodeurl/:platform/:purpose") {
		ctl.apiPrepare("")

		return
	}

	if ctl.isPutRequest() {
		ctl.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title logout
// @Description community manager logout
// @Tags AuthOnCodePlatform
// @Accept json
// @Success 202 {object} controllers.respData
// @router / [put]
func (ctl *AuthController) Logout() {
	action := "community manager logout"

	ctl.logout()

	ctl.sendSuccessResp(action + " successfully")
}

// @Title Callback
// @Description callback of authentication by oauth2
// @Tags AuthOnCodePlatform
// @Accept json
// @Param	:platform	path 	string		true		"gitee/github"
// @Param	:purpose	path 	string		true		"purpose: login, sign"
// @Failure 400 auth_failed:               authenticated on code platform failed
// @Failure 401 unsupported_code_platform: unsupported code platform
// @Failure 402 refuse_to_authorize_email: the user refused to access his/her email
// @Failure 403 no_public_email:           no public email
// @Failure 500 system_error:              system error
// @router /:platform/:purpose [get]
func (ctl *AuthController) Callback() {
	purpose := ctl.GetString(":purpose")
	platform := ctl.GetString(":platform")
	authHelper, ok := platformAuth.Auth[purpose]
	if !ok {
		return
	}

	rs := func(errCode string, reason error) {
		ctl.setCookies(map[string]string{"error_code": errCode, "error_msg": reason.Error()})

		ctl.redirect(authHelper.WebRedirectDir(false))
	}

	if ctl.GetString("state") != authURLState {
		rs(errSystemError, fmt.Errorf("unkown state"))
		return
	}

	if err := ctl.GetString("error"); err != "" {
		rs(errAuthFailed, fmt.Errorf("%s, %s", err, ctl.GetString("error_description")))
		return
	}

	cp, err := authHelper.GetAuthInstance(platform)
	if err != nil {
		rs(errUnsupportedCodePlatform, err)
		return
	}

	// gitee don't pass the scope paramter
	token, err := cp.GetToken(ctl.GetString("code"), ctl.GetString("scope"))
	if err != nil {
		rs(errSystemError, err)
		return
	}

	if purpose != platformAuth.AuthApplyToLogin {
		rs(errSystemError, errors.New("unknown purpose"))

		return
	}

	pl, ec, err := ctl.genACPayload(platform, token)
	if err != nil {
		rs(ec, err)
		return
	}

	at, err := ctl.newApiToken(PermissionOwnerOfOrg, pl)
	if err != nil {
		rs(errSystemError, err)
		return
	}

	ctl.setToken(at)
	ctl.redirect(authHelper.WebRedirectDir(true))
}

func (ctl *AuthController) genACPayload(platform, platformToken string) (*acForCodePlatformPayload, string, error) {
	pt, err := platforms.NewPlatform(platformToken, "", platform)
	if err != nil {
		return nil, errSystemError, err
	}

	// user
	user, err := pt.GetUser()
	if err != nil {
		return nil, errSystemError, err
	}

	// orgs
	orgs, err := pt.ListOrg()
	if err != nil {
		return nil, errSystemError, err
	}
	if len(orgs) == 0 {
		return nil, errNoOrg, errors.New("no org")
	}

	return &acForCodePlatformPayload{
		User:     user,
		Platform: platform,
		Orgs:     orgs,
	}, "", nil
}

// @Title AuthCodeURL
// @Description get authentication code url
// @Tags AuthOnCodePlatform
// @Accept json
// @Param	:platform	path 	string		true		"gitee/github"
// @Param	:purpose	path 	string		true		"purpose: login, sign"
// @Success 200 {object} map
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 unsupported_code_platform:  unsupported code platform
// @Failure 402 unkown_purpose_for_auth:    unknown purpose parameter
// @router /authcodeurl/:platform/:purpose [get]
func (ctl *AuthController) AuthCodeURL() {
	action := "fetch auth code url of gitee/github"

	authHelper, ok := platformAuth.Auth[ctl.GetString(":purpose")]
	if !ok {
		ctl.sendFailedResponse(400, errUnkownPurposeForAuth, fmt.Errorf("unkonw purpose"), action)
		return
	}

	cp, err := authHelper.GetAuthInstance(ctl.GetString(":platform"))
	if err != nil {
		ctl.sendFailedResponse(400, errUnsupportedCodePlatform, err, action)
		return
	}

	ctl.sendSuccessResp(map[string]string{
		"url": cp.GetAuthCodeURL(authURLState),
	})
}

type acForCodePlatformPayload struct {
	User     string   `json:"user"`
	Platform string   `json:"platform"`
	Orgs     []string `json:"orgs"`
}

func (pl *acForCodePlatformPayload) isOwnerOfLink(link string) *failedApiResult {
	v, err := models.GetLink(link)
	if err != nil {
		if err.IsErrorOf(models.ErrNoLink) {
			return newFailedApiResult(400, errUnknownLink, err)
		}

		return parseModelError(err)
	}

	return pl.isOwnerOfOrg(v.Platform, v.OrgID)
}

func (pl *acForCodePlatformPayload) isOwnerOfOrg(platform, org string) *failedApiResult {
	if pl.Platform == platform {
		for _, v := range pl.Orgs {
			if v == org {
				return nil
			}
		}
	}

	return newFailedApiResult(400, errNotYoursOrg, fmt.Errorf("not the org of owner"))
}
