package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/util"
)

type failedResult struct {
	reason     error
	errCode    string
	statusCode int
}

func newFailedResult(statusCode int, errCode string, err error) *failedResult {
	return &failedResult{
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

func (this *baseController) sendFailedResultAsResp(fr *failedResult, doWhat string) {
	this.sendFailedResponse(fr.statusCode, fr.errCode, fr.reason, doWhat)
}

func (this *baseController) sendFailedResponse(statusCode int, errCode string, reason error, doWhat string) {
	if statusCode == 0 {
		statusCode, errCode = buildStatusAndErrCode(statusCode, errCode, reason)
		if statusCode >= 500 {
			beego.Error(fmt.Sprintf("Failed to %s, errCode: %s, err: %s", doWhat, errCode, reason.Error()))

			reason = fmt.Errorf("System error")
			errCode = util.ErrSystemError
		}
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

func (this *baseController) refreshAccessToken() (string, error) {
	ac, err := this.getAccessController()
	if err != nil {
		return "", err
	}
	return ac.RefreshToken(conf.AppConfig.APITokenExpiry, conf.AppConfig.APITokenKey)
}

func (this *baseController) getAccessController() (*accessController, error) {
	ac, ok := this.Data[apiAccessController]
	if !ok {
		return nil, fmt.Errorf("no access controller")
	}

	if v, ok := ac.(accessController); ok {
		return &v, nil
	}

	return nil, fmt.Errorf("can't convert to access controller instance")
}

func (this *baseController) tokenPayloadOfCodePlatform() (*acForCodePlatformPayload, error) {
	ac, err := this.getAccessController()
	if err != nil {
		return nil, err
	}

	cpa, ok := ac.Payload.(*acForCodePlatformPayload)
	if !ok {
		return nil, fmt.Errorf("invalid token payload")
	}

	return cpa, nil
}

func (this *baseController) tokenPayloadOfCorpManager() (*acForCorpManagerPayload, error) {
	ac, err := this.getAccessController()
	if err != nil {
		return nil, err
	}

	pl, ok := ac.Payload.(*acForCorpManagerPayload)
	if !ok {
		return nil, fmt.Errorf("invalid token payload")
	}

	return pl, nil
}

func (this *baseController) fetchInputPayload(info interface{}) error {
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, info); err != nil {
		return fmt.Errorf("invalid input payload: %s", err.Error())
	}
	return nil
}

func (this *baseController) checkPathParameter() error {
	rp := this.routerPattern()
	if rp == "" {
		return nil
	}

	items := strings.Split(rp, "/")
	for _, item := range items {
		if strings.HasPrefix(item, ":") && this.GetString(item) == "" {
			return fmt.Errorf("missing path parameter:%s", item)
		}
	}

	return nil
}

func (this *baseController) routerPattern() string {
	v, ok := this.Data["RouterPattern"]
	if ok {
		return v.(string)
	}
	return ""
}

func (this *baseController) apiPrepare(permission string) {
	if err := this.checkPathParameter(); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, "")
		this.StopRun()
	}

	if permission != "" {
		if v := this.checkApiReqToken(permission); v != nil {
			this.sendFailedResponse(v.statusCode, v.errCode, v.reason, "")
			this.StopRun()
		}
	}
}

func (this *baseController) checkApiReqToken(permission string) *failedResult {
	token := this.apiReqHeader(headerToken)
	if token == "" {
		return &failedResult{
			statusCode: 401,
			errCode:    util.ErrMissingToken,
			reason:     fmt.Errorf("no token passed"),
		}
	}

	var acp interface{}

	switch permission {
	case PermissionOwnerOfOrg:
		acp = &acForCodePlatformPayload{}
	case PermissionIndividualSigner:
		acp = &acForCodePlatformPayload{}
	case PermissionCorporAdmin:
		acp = &acForCorpManagerPayload{}
	case PermissionEmployeeManager:
		acp = &acForCorpManagerPayload{}
	}

	ac := &accessController{Payload: acp}

	if err := ac.ParseToken(token, conf.AppConfig.APITokenKey); err != nil {
		return &failedResult{
			statusCode: 401,
			errCode:    util.ErrUnknownToken,
			reason:     err,
		}
	}

	if err := ac.Verify([]string{permission}); err != nil {
		return &failedResult{
			statusCode: 403,
			errCode:    util.ErrInvalidToken,
			reason:     err,
		}
	}

	this.Data[apiAccessController] = *ac
	return nil
}

func (this *baseController) apiReqHeader(h string) string {
	return this.Ctx.Input.Header(h)
}

func (this *baseController) getRequestMethod() string {
	return this.Ctx.Request.Method
}

func (this *baseController) readInputFile(fileName string) ([]byte, *failedResult) {
	f, _, err := this.GetFile(fileName)
	if err != nil {
		return nil, newFailedResult(400, util.ErrInvalidParameter, err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, newFailedResult(500, util.ErrSystemError, err)
	}
	return data, nil
}
