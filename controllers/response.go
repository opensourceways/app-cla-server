package controllers

import (
	"fmt"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/models"
)

func (ctl *baseController) sendResponse(action string, body interface{}, statusCode int) {
	if statusCode != 0 {
		// if success, don't set status code, otherwise the header set in ctl.ServeJSON
		// will not work. The reason maybe the same as above.
		ctl.Ctx.ResponseWriter.WriteHeader(statusCode)
	}

	ctl.Data["json"] = respData{
		Data: body,
	}

	ctl.ServeJSON()

	ctl.operationLog(action, statusCode)
}

func (ctl *baseController) sendSuccessResp(action string, body interface{}) {
	ctl.sendResponse(action, body, 0)
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

	ctl.sendResponse(action, d, statusCode)
}

func (ctl *baseController) operationLog(action string, statusCode int) {
	user := ""
	if ac, err := ctl.getAccessController(); err == nil {
		user = ac.getUser()
	}

	if user != "" {
		logs.Info("%s, %s, %d", user, action, statusCode)
	}
}

func (ctl *baseController) addOperationLog(user, action string, statusCode int) {
	logs.Info("%s, %s, %d", user, action, statusCode)
}
