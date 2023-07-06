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

func (this *AuthController) Prepare() {
	this.stopRunIfSignSerivceIsUnabled()

	b := strings.HasSuffix(this.routerPattern(), "/authcodeurl/:platform/:purpose")
	if b || this.isPostRequest() {
		this.apiPrepare("")
	}
}

// @Title Callback
// @Description callback of authentication by oauth2
// @Param	:platform	path 	string		true		"gitee/github"
// @Param	:purpose	path 	string		true		"purpose: login, sign"
// @Failure 400 auth_failed:               authenticated on code platform failed
// @Failure 401 unsupported_code_platform: unsupported code platform
// @Failure 402 refuse_to_authorize_email: the user refused to access his/her email
// @Failure 403 no_public_email:           no public email
// @Failure 500 system_error:              system error
// @router /:platform/:purpose [get]
func (this *AuthController) Callback() {
	purpose := this.GetString(":purpose")
	platform := this.GetString(":platform")
	authHelper, ok := platformAuth.Auth[purpose]
	if !ok {
		return
	}

	rs := func(errCode string, reason error) {
		this.setCookies(map[string]string{"error_code": errCode, "error_msg": reason.Error()})

		this.redirect(authHelper.WebRedirectDir(false))
	}

	if this.GetString("state") != authURLState {
		rs(errSystemError, fmt.Errorf("unkown state"))
		return
	}

	if err := this.GetString("error"); err != "" {
		rs(errAuthFailed, fmt.Errorf("%s, %s", err, this.GetString("error_description")))
		return
	}

	cp, err := authHelper.GetAuthInstance(platform)
	if err != nil {
		rs(errUnsupportedCodePlatform, err)
		return
	}

	// gitee don't pass the scope paramter
	token, err := cp.GetToken(this.GetString("code"), this.GetString("scope"))
	if err != nil {
		rs(errSystemError, err)
		return
	}

	if purpose != platformAuth.AuthApplyToLogin {
		rs(errSystemError, errors.New("unknown purpose"))

		return
	}

	pl, ec, err := this.genACPayload(platform, token)
	if err != nil {
		rs(ec, err)
		return
	}

	at, err := this.newApiToken(PermissionOwnerOfOrg, pl)
	if err != nil {
		rs(errSystemError, err)
		return
	}

	this.setToken(at)
	this.redirect(authHelper.WebRedirectDir(true))
}

func (this *AuthController) genACPayload(platform, platformToken string) (*acForCodePlatformPayload, string, error) {
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
// @Param	:platform	path 	string		true		"gitee/github"
// @Param	:purpose	path 	string		true		"purpose: login, sign"
// @Success 200 {object} map
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 unsupported_code_platform:  unsupported code platform
// @Failure 402 unkown_purpose_for_auth:    unknown purpose parameter
// @router /authcodeurl/:platform/:purpose [get]
func (this *AuthController) AuthCodeURL() {
	action := "fetch auth code url of gitee/github"

	authHelper, ok := platformAuth.Auth[this.GetString(":purpose")]
	if !ok {
		this.sendFailedResponse(400, errUnkownPurposeForAuth, fmt.Errorf("unkonw purpose"), action)
		return
	}

	cp, err := authHelper.GetAuthInstance(this.GetString(":platform"))
	if err != nil {
		this.sendFailedResponse(400, errUnsupportedCodePlatform, err, action)
		return
	}

	this.sendSuccessResp(map[string]string{
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
