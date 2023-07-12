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

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

const (
	csrfToken   = "csrf_token"
	accessToken = "access_token"
)

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
}

func (ctl *baseController) sendResponse(body interface{}, statusCode int) {
	if statusCode != 0 {
		// if success, don't set status code, otherwise the header set in ctl.ServeJSON
		// will not work. The reason maybe the same as above.
		ctl.Ctx.ResponseWriter.WriteHeader(statusCode)
	}

	ctl.Data["json"] = respData{
		Data: body,
	}

	ctl.ServeJSON()
}

func (ctl *baseController) sendSuccessResp(body interface{}) {
	ctl.sendResponse(body, 0)
}

func (ctl *baseController) newFuncForSendingFailedResp(action string) func(fr *failedApiResult) {
	return func(fr *failedApiResult) {
		ctl.sendFailedResponse(fr.statusCode, fr.errCode, fr.reason, action)
	}
}

func (ctl *baseController) sendModelErrorAsResp(err models.IModelError, action string) {
	ctl.sendFailedResultAsResp(parseModelError(err), action)
}

func (ctl *baseController) sendFailedResultAsResp(fr *failedApiResult, action string) {
	ctl.sendFailedResponse(fr.statusCode, fr.errCode, fr.reason, action)
}

func (ctl *baseController) sendFailedResponse(statusCode int, errCode string, reason error, action string) {
	if statusCode >= 500 {
		logs.Error(fmt.Sprintf("Failed to %s, errCode: %s, err: %s", action, errCode, reason.Error()))

		errCode = errSystemError
		reason = fmt.Errorf("system error")
	}

	d := errMsg{
		ErrCode: fmt.Sprintf("cla.%s", errCode),
		ErrMsg:  reason.Error(),
	}

	ctl.sendResponse(d, statusCode)
}

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

func (ctl *baseController) fetchInputPayload(info interface{}) *failedApiResult {
	return fetchInputPayloadData(ctl.Ctx.Input.RequestBody, info)
}

func (ctl *baseController) fetchInputPayloadFromFormData(info interface{}) *failedApiResult {
	input := []byte(ctl.Ctx.Request.FormValue("data"))
	return fetchInputPayloadData(input, info)
}

func (ctl *baseController) checkPathParameter() *failedApiResult {
	rp := ctl.routerPattern()
	if rp == "" {
		return nil
	}

	items := strings.Split(rp, "/")
	for _, item := range items {
		if strings.HasPrefix(item, ":") && ctl.GetString(item) == "" {
			return newFailedApiResult(
				400, errMissingURLPathParameter,
				fmt.Errorf("missing path parameter:%s", item))
		}
	}

	return nil
}

func (ctl *baseController) routerPattern() string {
	if v, ok := ctl.Data["RouterPattern"]; ok {
		return v.(string)
	}
	return ""
}

func (ctl *baseController) apiPrepare(permission string) {
	if permission != "" {
		ctl.apiPrepareWithAC(
			ctl.newAccessController(permission),
			[]string{permission},
		)
	} else {
		ctl.apiPrepareWithAC(nil, nil)
	}
}

func (ctl *baseController) apiPrepareWithAC(ac *accessController, permission []string) {
	if fr := ctl.checkPathParameter(); fr != nil {
		ctl.sendFailedResultAsResp(fr, "")
		ctl.StopRun()
	}

	if ac != nil && len(permission) != 0 {
		if fr := ctl.checkApiReqToken(ac, permission); fr != nil {
			ctl.sendFailedResultAsResp(fr, "")
			ctl.StopRun()
		}

		ctl.Data[apiAccessController] = *ac
	}
}

func (ctl *baseController) newAccessController(permission string) *accessController {
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

func (ctl *baseController) checkApiReqToken(ac *accessController, permission []string) *failedApiResult {
	token, fr := ctl.getToken()
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

	ctl.setToken(newToken)

	if err := json.Unmarshal(v, ac); err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}

	addr, fr := ctl.getRemoteAddr()
	if fr != nil {
		return fr
	}

	if err := ac.verify(permission, addr); err != nil {
		return newFailedApiResult(403, errUnauthorizedToken, err)
	}

	return nil
}

func (ctl *baseController) getAccessController() (*accessController, *failedApiResult) {
	ac, ok := ctl.Data[apiAccessController]
	if !ok {
		return nil, newFailedApiResult(500, errSystemError, fmt.Errorf("no access controller"))
	}

	if v, ok := ac.(accessController); ok {
		return &v, nil
	}

	return nil, newFailedApiResult(500, errSystemError, fmt.Errorf("can't convert to access controller instance"))
}

func (ctl *baseController) apiReqHeader(h string) string {
	return ctl.Ctx.Input.Header(h)
}

func (ctl *baseController) apiRequestMethod() string {
	return ctl.Ctx.Request.Method
}

func (ctl *baseController) isPostRequest() bool {
	return ctl.apiRequestMethod() == http.MethodPost
}

func (ctl *baseController) isPutRequest() bool {
	return ctl.apiRequestMethod() == http.MethodPut
}

func (ctl *baseController) isGetRequest() bool {
	return ctl.apiRequestMethod() == http.MethodGet
}

func (ctl *baseController) readInputFile(fileName string, maxSize int, fileType string) ([]byte, *failedApiResult) {
	if v := ctl.Ctx.Request.ContentLength; v <= 0 || v > int64(maxSize) {
		return nil, newFailedApiResult(400, errTooBigPDF, fmt.Errorf("big pdf file"))
	}

	f, _, err := ctl.GetFile(fileName)
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

	if !util.CheckContentType(data, fileType) {
		return nil, newFailedApiResult(400, errWrongFileType, fmt.Errorf("unsupported file type"))
	}

	return data, nil
}

func (ctl *baseController) downloadFile(path string) {
	ctl.Ctx.Output.Download(path)
}

func (ctl *baseController) redirect(webRedirectDir string) {
	http.Redirect(
		ctl.Ctx.ResponseWriter, ctl.Ctx.Request, webRedirectDir, http.StatusFound,
	)
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
	for _, item := range strings.Split(ips, ", ") {
		if net.ParseIP(item) != nil {
			return item, nil
		}
	}

	return "", newFailedApiResult(400, errCanNotFetchClientIP, fmt.Errorf("can not fetch client ip"))
}

func (ctl *baseController) logout() {
	if t := ctl.Ctx.GetCookie(accessToken); t != "" {
		models.RemoveAccessToken(t)
	}
	ctl.setToken(models.AccessToken{})
}
