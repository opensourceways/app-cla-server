package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/models"
)

const accessToken = "access_token"

type failedApiResult struct {
	reason     error
	errCode    string
	statusCode int
}

func newFailedApiResult(statusCode int, errCode string, err error) *failedApiResult {
	return &failedApiResult{
		statusCode: statusCode,
		errCode:    errCode,
		reason:     err,
	}
}

type baseController struct {
	beego.Controller

	ac *accessController
}

func (this *baseController) sendResponse(body interface{}, statusCode int) {
	if statusCode != 0 {
		// if success, don't set status code, otherwise the header set in this.ServeJSON
		// will not work. The reason maybe the same as above.
		this.Ctx.ResponseWriter.WriteHeader(statusCode)
	}

	this.Data["json"] = respData{
		Data: body,
	}

	this.ServeJSON()
}

func (this *baseController) sendSuccessResp(body interface{}) {
	this.sendResponse(body, 0)
}

func (this *baseController) newFuncForSendingFailedResp(action string) func(fr *failedApiResult) {
	return func(fr *failedApiResult) {
		this.sendFailedResponse(fr.statusCode, fr.errCode, fr.reason, action)
	}
}

func (this *baseController) sendModelErrorAsResp(err models.IModelError, action string) {
	this.sendFailedResultAsResp(parseModelError(err), action)
}

func (this *baseController) sendFailedResultAsResp(fr *failedApiResult, action string) {
	this.sendFailedResponse(fr.statusCode, fr.errCode, fr.reason, action)
}

func (this *baseController) sendFailedResponse(statusCode int, errCode string, reason error, action string) {
	if statusCode >= 500 {
		logs.Error(fmt.Sprintf("Failed to %s, errCode: %s, err: %s", action, errCode, reason.Error()))

		errCode = errSystemError
		reason = fmt.Errorf("system error")
	}

	d := errMsg{
		ErrCode: fmt.Sprintf("cla.%s", errCode),
		ErrMsg:  reason.Error(),
	}

	this.sendResponse(d, statusCode)
}

func (this *baseController) newApiToken(permission string, pl interface{}) (models.AccessToken, error) {
	addr, fr := this.getRemoteAddr()
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

func (this *baseController) tokenPayloadBasedOnCodePlatform() (*acForCodePlatformPayload, *failedApiResult) {
	ac, fr := this.getAccessController()
	if fr != nil {
		return nil, fr
	}

	if pl, ok := ac.Payload.(*acForCodePlatformPayload); ok {
		return pl, nil
	}
	return nil, newFailedApiResult(500, errSystemError, fmt.Errorf("invalid token payload"))
}

func (this *baseController) tokenPayloadBasedOnCorpManager() (*acForCorpManagerPayload, *failedApiResult) {
	ac, fr := this.getAccessController()
	if fr != nil {
		return nil, fr
	}

	if pl, ok := ac.Payload.(*acForCorpManagerPayload); ok {
		return pl, nil
	}
	return nil, newFailedApiResult(500, errSystemError, fmt.Errorf("invalid token payload"))
}

func (this *baseController) fetchInputPayload(info interface{}) *failedApiResult {
	return fetchInputPayloadData(this.Ctx.Input.RequestBody, info)
}

func (this *baseController) fetchInputPayloadFromFormData(info interface{}) *failedApiResult {
	input := []byte(this.Ctx.Request.FormValue("data"))
	return fetchInputPayloadData(input, info)
}

func (this *baseController) checkPathParameter() *failedApiResult {
	rp := this.routerPattern()
	if rp == "" {
		return nil
	}

	items := strings.Split(rp, "/")
	for _, item := range items {
		if strings.HasPrefix(item, ":") && this.GetString(item) == "" {
			return newFailedApiResult(
				400, errMissingURLPathParameter,
				fmt.Errorf("missing path parameter:%s", item))
		}
	}

	return nil
}

func (this *baseController) routerPattern() string {
	if v, ok := this.Data["RouterPattern"]; ok {
		return v.(string)
	}
	return ""
}

func (this *baseController) apiPrepare(permission string) {
	if permission != "" {
		this.apiPrepareWithAC(
			this.newAccessController(permission),
			[]string{permission},
		)
	} else {
		this.apiPrepareWithAC(nil, nil)
	}
}

func (this *baseController) apiPrepareWithAC(ac *accessController, permission []string) {
	if fr := this.checkPathParameter(); fr != nil {
		this.sendFailedResultAsResp(fr, "")
		this.StopRun()
	}

	if ac != nil && permission != nil {
		if fr := this.checkApiReqToken(ac, permission); fr != nil {
			this.sendFailedResultAsResp(fr, "")
			this.StopRun()
		}

		this.Data[apiAccessController] = *ac
	}
}

func (this *baseController) newAccessController(permission string) *accessController {
	var acp interface{}

	switch permission {
	case PermissionOwnerOfOrg:
		acp = &acForCodePlatformPayload{}
	case PermissionCorpAdmin:
		acp = &acForCorpManagerPayload{}
	case PermissionEmployeeManager:
		acp = &acForCorpManagerPayload{}
	}

	return &accessController{Payload: acp}
}

func (this *baseController) checkApiReqToken(ac *accessController, permission []string) *failedApiResult {
	token, fr := this.getToken()
	if fr != nil {
		return fr
	}

	newToken, v, err := models.ValidateAndRefreshAccessToken(token)

	if err != nil {
		if err.IsErrorOf(models.ErrInvalidToken) {
			return newFailedApiResult(401, errUnknownToken, err)
		}

		return newFailedApiResult(500, errSystemError, err)
	}

	this.setToken(newToken)

	if err := json.Unmarshal(v, ac); err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}

	addr, fr := this.getRemoteAddr()
	if fr != nil {
		return fr
	}

	if err := ac.verify(permission, addr); err != nil {
		return newFailedApiResult(403, errUnauthorizedToken, err)
	}

	return nil
}

func (this *baseController) getAccessController() (*accessController, *failedApiResult) {
	ac, ok := this.Data[apiAccessController]
	if !ok {
		return nil, newFailedApiResult(500, errSystemError, fmt.Errorf("no access controller"))
	}

	if v, ok := ac.(accessController); ok {
		return &v, nil
	}

	return nil, newFailedApiResult(500, errSystemError, fmt.Errorf("can't convert to access controller instance"))
}

func (this *baseController) apiReqHeader(h string) string {
	return this.Ctx.Input.Header(h)
}

func (this *baseController) apiRequestMethod() string {
	return this.Ctx.Request.Method
}

func (this *baseController) isPostRequest() bool {
	return this.apiRequestMethod() == http.MethodPost
}

func (this *baseController) readInputFile(fileName string, maxSize int) ([]byte, *failedApiResult) {
	if v := this.Ctx.Request.ContentLength; v <= 0 || v > int64(maxSize) {
		return nil, newFailedApiResult(400, errTooBigPDF, fmt.Errorf("big pdf file"))
	}

	f, _, err := this.GetFile(fileName)
	if err != nil {
		return nil, newFailedApiResult(400, errReadingFile, err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, newFailedApiResult(500, errSystemError, err)
	}

	if len(data) > maxSize {
		return nil, newFailedApiResult(400, errTooBigPDF, fmt.Errorf("big pdf file"))
	}

	return data, nil
}

func (this *baseController) downloadFile(path string) {
	this.Ctx.Output.Download(path)
}

func (this *baseController) redirect(webRedirectDir string) {
	http.Redirect(
		this.Ctx.ResponseWriter, this.Ctx.Request, webRedirectDir, http.StatusFound,
	)
}

func (this *baseController) setCookies(value map[string]string) {
	cfg := config.AppConfig.APIConfig

	// TODO check the property of cookie
	for k, v := range value {
		this.Ctx.SetCookie(k, v, cfg.CookieTimeout, "/")
	}
}

func (this *baseController) getToken() (t models.AccessToken, fr *failedApiResult) {
	if t.CSRF = this.apiReqHeader(headerToken); t.CSRF == "" {
		fr = newFailedApiResult(401, errMissingToken, fmt.Errorf("no token passed"))

		return
	}

	if t.Id = this.Ctx.GetCookie(accessToken); t.Id == "" {
		fr = newFailedApiResult(401, errMissingToken, fmt.Errorf("no token passed"))
	}

	return
}

func (this *baseController) setToken(t models.AccessToken) {
	this.setCookies(map[string]string{"csrf": t.CSRF})

	// TODO samesite?
	cfg := config.AppConfig.APIConfig

	this.Ctx.SetCookie(
		accessToken, t.Id, cfg.CookieTimeout, "/", cfg.CookieDomain, true, true,
	)
}

func (this *baseController) getRemoteAddr() (string, *failedApiResult) {
	ips := this.Ctx.Request.Header.Get("x-forwarded-for")
	for _, item := range strings.Split(ips, ", ") {
		if net.ParseIP(item) != nil {
			return item, nil
		}
	}

	return "", newFailedApiResult(400, errCanNotFetchClientIP, fmt.Errorf("can not fetch client ip"))
}

func (this *baseController) stopRunIfSignSerivceIsUnabled() {
	if config.AppConfig == nil {
		this.StopRun()
	}
}
