package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/astaxie/beego"

	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
	"github.com/opensourceways/app-cla-server/code-platform-auth/platforms"
	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
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
	purpose := this.GetString(":purpose")
	platform := this.GetString(":platform")
	authHelper, ok := platformAuth.Auth[purpose]
	if !ok {
		return
	}

	if this.GetString("state") != authURLState {
		return
	}

	rs := func(errCode string, reason error) {
		rspOnAuthFailed(&this.Controller, authHelper.WebRedirectDir(false), errCode, reason)
	}

	if err := this.GetString("error"); err != "" {
		rs(util.ErrAuthFailed, fmt.Errorf("%s, %s", err, this.GetString("error_description")))
		return
	}

	cp, err := authHelper.GetAuthInstance(platform)
	if err != nil {
		rs(util.ErrNotSupportedPlatform, err)
		return
	}

	// gitee don't pass the scope paramter
	token, err := cp.GetToken(this.GetString("code"), this.GetString("scope"))
	if err != nil {
		rs(util.ErrSystemError, err)
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

	at, err := this.newAccessToken(permission, pl)
	if err != nil {
		rs(util.ErrSystemError, err)
		return
	}

	cookies := map[string]string{"access_token": at, "platform_token": token}
	if permission == PermissionIndividualSigner {
		cookies["sign_user"] = pl.User
		cookies["sign_email"] = pl.Email
	}
	setCookies(&this.Controller, cookies)

	http.Redirect(
		this.Ctx.ResponseWriter, this.Ctx.Request,
		authHelper.WebRedirectDir(true), http.StatusFound,
	)
}

func (this *AuthController) genACPayload(platform, permission, platformToken string) (*acForCodePlatformPayload, string, error) {
	pt, err := platforms.NewPlatform(platformToken, "", platform)
	if err != nil {
		return nil, util.ErrNotSupportedPlatform, err
	}

	orgm := map[string]bool{}
	links := map[string]dbmodels.OrgInfo{}
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
			if strings.Index(err.Error(), "401") >= 0 {
				return nil, util.ErrUnauthorized, err
			}
			return nil, util.ErrSystemError, err
		}
	}

	user, err := pt.GetUser()
	if err != nil {
		return nil, util.ErrSystemError, err
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

func (this *AuthController) newAccessToken(permission string, pl *acForCodePlatformPayload) (string, error) {
	ac := &accessController{
		Expiry:     util.Expiry(conf.AppConfig.APITokenExpiry),
		Permission: permission,
		Payload:    pl,
	}

	return ac.NewToken(conf.AppConfig.APITokenKey)
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

	authHelper, ok := platformAuth.Auth[this.GetString(":purpose")]
	if !ok {
		reason = fmt.Errorf("unkonw purpose")
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	cp, err := authHelper.GetAuthInstance(this.GetString(":platform"))
	if err != nil {
		reason = err
		errCode = util.ErrNotSupportedPlatform
		statusCode = 400
		return
	}

	body = map[string]string{
		"url": cp.GetAuthCodeURL(authURLState),
	}
}

type acForCodePlatformPayload struct {
	User          string `json:"user"`
	Email         string `json:"email"`
	Platform      string `json:"platform"`
	PlatformToken string `json:"platform_token"`

	Orgs  map[string]bool             `json:"orgs"`
	Links map[string]dbmodels.OrgInfo `json:links`
}

func (this *acForCodePlatformPayload) orgInfo(linkID string) *dbmodels.OrgInfo {
	if this.Links == nil {
		return nil
	}

	if v, ok := this.Links[linkID]; ok {
		return &v
	}
	return nil
}
func (this *acForCodePlatformPayload) isOwnerOfLink(link string) *failedResult {
	if this.Links == nil {
		this.Links = map[string]dbmodels.OrgInfo{}
	}

	if _, ok := this.Links[link]; ok {
		return nil
	}

	orgInfo, err := models.GetOrgOfLink(link)
	if err != nil {
		// TODO check if link is not exist
	}

	if err := this.isOwnerOfOrg(orgInfo.OrgID); err != nil {
		return err
	}

	this.Links[link] = *orgInfo
	return nil
}

func (this *acForCodePlatformPayload) isOwnerOfOrg(org string) *failedResult {
	if this.Orgs == nil {
		this.Orgs = map[string]bool{}
	}

	if this.Orgs[org] {
		return nil
	}

	p, err := platforms.NewPlatform(this.PlatformToken, "", this.Platform)
	if err != nil {
		return newFailedResult(400, util.ErrInvalidParameter, err)
	}

	if b, err := p.IsOrgExist(org); err != nil {
		// TODO token expiry
		return newFailedResult(500, util.ErrSystemError, err)
	} else if !b {
		return newFailedResult(400, util.ErrNotYoursOrg, fmt.Errorf("not the org of owner"))
	}

	this.Orgs[org] = true
	return nil
}
