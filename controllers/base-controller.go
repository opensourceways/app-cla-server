package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
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

	ac *accessController
}

func (this *baseController) sendResponse(body interface{}, statusCode int) {
	if token, err := this.refreshAccessToken(); err == nil {
		// this code must run before `this.Ctx.ResponseWriter.WriteHeader`
		// otherwise the header can't be set successfully.
		// The reason is relevant to the variable of 'Response.Started' at
		// beego/context/context.go
		this.Ctx.Output.Header(headerToken, token)
	}

	if statusCode != 0 {
		// if success, don't set status code, otherwise the header set in this.ServeJSON
		// will not work. The reason maybe the same as above.
		this.Ctx.ResponseWriter.WriteHeader(statusCode)
	}

	this.Data["json"] = struct {
		Data interface{} `json:"data"`
	}{
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
		beego.Error(fmt.Sprintf("Failed to %s, errCode: %s, err: %s", action, errCode, reason.Error()))

		errCode = errSystemError
		reason = fmt.Errorf("System error")
	}

	d := struct {
		ErrCode string `json:"error_code"`
		ErrMsg  string `json:"error_message"`
	}{
		ErrCode: fmt.Sprintf("cla.%s", errCode),
		ErrMsg:  reason.Error(),
	}

	this.sendResponse(d, statusCode)
}

func (this *baseController) newApiToken(permission string, pl interface{}) (string, error) {
	ac := &accessController{
		Expiry:     util.Expiry(config.AppConfig.APITokenExpiry),
		Permission: permission,
		Payload:    pl,
	}

	return ac.NewToken(config.AppConfig.APITokenKey)
}

func (this *baseController) refreshAccessToken() (string, *failedApiResult) {
	ac, fr := this.getAccessController()
	if fr != nil {
		return "", fr
	}

	token, err := ac.RefreshToken(config.AppConfig.APITokenExpiry, config.AppConfig.APITokenKey)
	if err == nil {
		return token, nil
	}
	return "", newFailedApiResult(500, errSystemError, err)
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
	return this.fetchInputPayloadData(&this.Ctx.Input.RequestBody, info)
}

func (this *baseController) fetchInputPayloadFromFormData(info interface{}) *failedApiResult {
	input := []byte(this.Ctx.Request.FormValue("data"))
	return this.fetchInputPayloadData(&input, info)
}

func (this *baseController) fetchInputPayloadData(input *[]byte, info interface{}) *failedApiResult {
	if err := json.Unmarshal(*input, info); err != nil {
		return newFailedApiResult(
			400, errParsingApiBody, fmt.Errorf("invalid input payload: %s", err.Error()),
		)
	}
	return nil
}

func (this *baseController) checkPathParameter() *failedApiResult {
	rp := this.routerPattern()
	if rp == "" {
		return nil
	}

	items := strings.Split(rp, "/")
	for _, item := range items {
		if strings.HasPrefix(item, ":") && this.GetString(item) == "" {
			return newFailedApiResult(400, errMissingParameter, fmt.Errorf("missing path parameter:%s", item))
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
	case PermissionIndividualSigner:
		acp = &acForCodePlatformPayload{}
	case PermissionCorpAdmin:
		acp = &acForCorpManagerPayload{}
	case PermissionEmployeeManager:
		acp = &acForCorpManagerPayload{}
	}

	return &accessController{Payload: acp}
}

func (this *baseController) checkApiReqToken(ac *accessController, permission []string) *failedApiResult {
	token := this.apiReqHeader(headerToken)
	if token == "" {
		return newFailedApiResult(401, errMissingToken, fmt.Errorf("no token passed"))
	}

	if err := ac.ParseToken(token, config.AppConfig.APITokenKey); err != nil {
		return newFailedApiResult(401, errUnknownToken, err)
	}

	if err := ac.Verify(permission); err != nil {
		return newFailedApiResult(403, errInvalidToken, err)
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

func (this *baseController) readInputFile(fileName string) ([]byte, *failedApiResult) {
	f, _, err := this.GetFile(fileName)
	if err != nil {
		return nil, newFailedApiResult(400, errReadingFile, err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, newFailedApiResult(500, errSystemError, err)
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
	for k, v := range value {
		this.Ctx.SetCookie(k, v, "3600", "/")
	}
}
