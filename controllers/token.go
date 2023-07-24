package controllers

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/opensourceways/app-cla-server/models"
)

const (
	csrfToken   = "csrf_token"
	accessToken = "access_token"
)

func (ctl *baseController) newApiToken(permission string, pl interface{}) (models.AccessToken, error) {
	addr, fr := ctl.getRemoteAddr()
	if fr != nil {
		return models.AccessToken{}, fr.reason
	}

	ac := &accessController{
		Payload:    pl,
		RemoteAddr: addr,
		Permission: permission,
	}

	v, err := json.Marshal(ac)
	if err != nil {
		return models.AccessToken{}, err
	}

	return models.NewAccessToken(v)
}

func (ctl *baseController) tokenPayloadBasedOnCodePlatform() (*acForCodePlatformPayload, *failedApiResult) {
	ac, fr := ctl.getAccessController()
	if fr != nil {
		return nil, fr
	}

	if pl, ok := ac.Payload.(*acForCodePlatformPayload); ok {
		return pl, nil
	}
	return nil, newFailedApiResult(500, errSystemError, fmt.Errorf("invalid token payload"))
}

func (ctl *baseController) tokenPayloadBasedOnCorpManager() (*acForCorpManagerPayload, *failedApiResult) {
	ac, fr := ctl.getAccessController()
	if fr != nil {
		return nil, fr
	}

	if pl, ok := ac.Payload.(*acForCorpManagerPayload); ok {
		return pl, nil
	}
	return nil, newFailedApiResult(500, errSystemError, fmt.Errorf("invalid token payload"))
}

func (ctl *baseController) setCookies(value map[string]string) {
	for k, v := range value {
		ctl.setCookie(k, v, false)
	}
}

func (ctl *baseController) setCookie(k, v string, httpOnly bool) {
	ctl.Ctx.SetCookie(
		k, v, config.CookieTimeout, "/", config.CookieDomain, true, httpOnly, "strict",
	)
}

func (ctl *baseController) getToken() (t models.AccessToken, fr *failedApiResult) {
	if t.CSRF = ctl.apiReqHeader(headerToken); t.CSRF == "" {
		fr = newFailedApiResult(401, errMissingToken, fmt.Errorf("no token passed"))

		return
	}

	if t.Id = ctl.Ctx.GetCookie(accessToken); t.Id == "" {
		fr = newFailedApiResult(401, errMissingToken, fmt.Errorf("no token passed"))
	}

	return
}

func (ctl *baseController) setToken(t models.AccessToken) {
	ctl.setCookie(csrfToken, t.CSRF, false)
	ctl.setCookie(accessToken, t.Id, true)
}

func (ctl *baseController) getRemoteAddr() (string, *failedApiResult) {
	ips := ctl.Ctx.Request.Header.Get("x-forwarded-for")
	logs.Info("x-forwarded-for value is: %s", ips)

	for _, item := range strings.Split(ips, ", ") {
		if net.ParseIP(item) != nil {
			logs.Info("x-forwarded-for value is: %s, remote addr: %s", ips, item)

			return item, nil
		}
	}

	logs.Info("x-forwarded-for value is: %s, no remote addr", ips)

	return "", newFailedApiResult(400, errCanNotFetchClientIP, fmt.Errorf("can not fetch client ip"))
}
