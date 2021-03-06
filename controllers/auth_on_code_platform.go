package controllers

import (
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
	if this.isPostRequest() ||
		strings.HasSuffix(this.routerPattern(), "/authcodeurl/:platform/:purpose") {

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
		this.setCookies(map[string]string{"error_code": errCode, "error_msg": reason.Error()}, false)
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

	ip, fr := this.getRemoteAddr()
	if fr != nil {
		rs(fr.errCode, fr.reason)
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

	permission := ""
	switch purpose {
	case platformAuth.AuthApplyToLogin:
		permission = PermissionOwnerOfOrg
	case platformAuth.AuthApplyToSign:
		permission = PermissionIndividualSigner
	}

	pl, ec, err := this.genACPayload(platform, permission, token)
	if err != nil {
		rs(ec, err)
		return
	}

	at, err := this.newApiToken(permission, ip, pl)
	if err != nil {
		rs(errSystemError, err)
		return
	}
	this.setCookies(map[string]string{apiAccessToken: at}, true)

	cookies := map[string]string{"action": purpose, "platform": platform}
	if permission == PermissionIndividualSigner {
		cookies["sign_user"] = pl.User
		cookies["sign_email"] = pl.Email
	}
	this.setCookies(cookies, false)

	this.redirect(authHelper.WebRedirectDir(true))
}

type userAccount struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

// @Title Auth
// @Description authentication by user's password of code platform
// @Param	:platform	path 	string				true	"gitee/github"
// @Param	body		body 	controllers.userAccount		true	"body for auth on code platform"
// @Success 201 {object} map
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 error_parsing_api_body:     parse payload of request failed
// @Failure 402 unsupported_code_platform: unsupported code platform
// @Failure 500 system_error:              system error
// @router /:platform [post]
func (this *AuthController) Auth() {
	action := "auth by pw"
	platform := this.GetString(":platform")

	var body userAccount

	if fr := this.fetchInputPayload(&body); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	ip, fr := this.getRemoteAddr()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	cp, err := platformAuth.Auth[platformAuth.AuthApplyToLogin].GetAuthInstance(platform)
	if err != nil {
		this.sendFailedResponse(400, errUnsupportedCodePlatform, err, action)
		return
	}

	token, err := cp.PasswordCredentialsToken(body.UserName, body.Password)
	if err != nil {
		this.sendFailedResponse(500, errSystemError, err, action)
		return
	}

	permission := PermissionOwnerOfOrg
	pl, ec, err := this.genACPayload(platform, permission, token)
	if err != nil {
		this.sendFailedResponse(500, ec, err, action)
		return
	}

	at, err := this.newApiToken(permission, ip, pl)
	if err != nil {
		this.sendFailedResponse(500, errSystemError, err, action)
		return
	}

	this.sendSuccessResp(map[string]string{apiAccessToken: at})
}

func (this *AuthController) genACPayload(platform, permission, platformToken string) (*acForCodePlatformPayload, string, error) {
	pt, err := platforms.NewPlatform(platformToken, "", platform)
	if err != nil {
		return nil, errSystemError, err
	}

	orgm := map[string]bool{}
	links := map[string]models.OrgInfo{}
	if permission == PermissionOwnerOfOrg {
		orgs, err := pt.ListOrg()
		if err == nil {
			for _, item := range orgs {
				orgm[item] = true
			}

			if r, err := models.ListLinks(platform, orgs); err == nil {
				for i := range r {
					links[r[i].LinkID] = r[i].OrgInfo
				}
			}
		}
	}

	email := ""
	if permission == PermissionIndividualSigner {
		if email, err = pt.GetAuthorizedEmail(); err != nil {
			if platforms.IsErrOfRefusedToAuthorizeEmail(err) {
				return nil, errRefuseToAuthorizeEmail, err
			}
			if platforms.IsErrOfNoPulicEmail(err) {
				return nil, errNoPublicEmail, err
			}
			return nil, errSystemError, err
		}
	}

	user, err := pt.GetUser()
	if err != nil {
		return nil, errSystemError, err
	}

	return &acForCodePlatformPayload{
		User:          user,
		Email:         email,
		Platform:      platform,
		PlatformToken: platformToken,
		Orgs:          orgm,
		Links:         links,
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
	User          string `json:"user"`
	Email         string `json:"email"`
	Platform      string `json:"platform"`
	PlatformToken string `json:"platform_token"`

	Orgs  map[string]bool           `json:"orgs"`
	Links map[string]models.OrgInfo `json:"links"`
}

func (this *acForCodePlatformPayload) orgInfo(linkID string) *models.OrgInfo {
	if this.Links == nil {
		return nil
	}

	if v, ok := this.Links[linkID]; ok {
		return &v
	}
	return nil
}

func (this *acForCodePlatformPayload) isOwnerOfLink(link string) *failedApiResult {
	if this.Links == nil {
		this.Links = map[string]models.OrgInfo{}
	}

	if _, ok := this.Links[link]; ok {
		return nil
	}

	orgInfo, err := models.GetOrgOfLink(link)
	if err != nil {
		if err.IsErrorOf(models.ErrNoLink) {
			return newFailedApiResult(400, errUnknownLink, err)
		}
		return parseModelError(err)
	}

	if err := this.isOwnerOfOrg(orgInfo.OrgID); err != nil {
		return err
	}

	this.Links[link] = *orgInfo
	return nil
}

func (this *acForCodePlatformPayload) isOwnerOfOrg(org string) *failedApiResult {
	if this.Orgs == nil {
		this.Orgs = map[string]bool{}
	}

	if this.Orgs[org] {
		return nil
	}

	this.refreshOrg()

	if !this.Orgs[org] {
		return newFailedApiResult(400, errNotYoursOrg, fmt.Errorf("not the org of owner"))
	}
	return nil
}

func (this *acForCodePlatformPayload) refreshOrg() {
	pt, err := platforms.NewPlatform(this.PlatformToken, "", this.Platform)
	if err != nil {
		return
	}

	// TODO token expiry
	orgs, err := pt.ListOrg()
	if err != nil {
		return
	}

	for _, item := range orgs {
		this.Orgs[item] = true
	}
}

func (this *acForCodePlatformPayload) hasRepo(org, repo string) (bool, *failedApiResult) {
	pt, err := platforms.NewPlatform(this.PlatformToken, "", this.Platform)
	if err != nil {
		return false, newFailedApiResult(400, errSystemError, err)
	}

	b, err := pt.HasRepo(org, repo)
	if err != nil {
		return false, newFailedApiResult(400, errSystemError, err)
	}

	return b, nil
}
