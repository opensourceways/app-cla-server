package controllers

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/smtpimpl"
)

type SMTPController struct {
	baseController
}

func (ctl *SMTPController) Prepare() {
	ctl.apiPrepare(PermissionOwnerOfOrg)
}

// @Title Verify
// @Description verify the email
// @Tags SMTP
// @Accept json
// @Param  body  body  models.EmailAuthorizationReq  true  "body for verifying the email"
// @Success 201 {object} controllers.respData
// @router /verify [post]
func (ctl *SMTPController) Verify() {
	action := "community manager verifies the email"

	var info models.EmailAuthorizationReq
	if fr := ctl.fetchInputPayloadFromFormData(&info); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	code, me := models.VerifySMTPEmail(&info)
	if me != nil {
		ctl.sendModelErrorAsResp(me, action)
		return
	}

	e := emailtmpl.EmailVerification{
		Code: code,
	}
	msg, err := e.GenEmailMsg()
	if err != nil {
		ctl.sendFailedResponse(500, errSystemError, err, action)

		return
	}

	msg.From = info.Email
	msg.To = []string{info.Email}
	msg.Subject = "CLA Email authorization verification code"

	if err = smtpimpl.SMTP().Send(info.Authorize, &msg); err != nil {
		ctl.sendFailedResponse(400, errInvalidEmailAuthCode, err, action)
	} else {
		ctl.sendSuccessResp(action, "successfully")
	}
}

// @Title Authorize
// @Description authorize the email
// @Tags SMTP
// @Accept json
// @Param  body  body  models.EmailAuthorization  true  "body for authorizing the email"
// @Success 201 {object} controllers.respData
// @router /authorize [post]
func (ctl *SMTPController) Authorize() {
	action := "community manager authorizes the email"

	var info models.EmailAuthorization
	if fr := ctl.fetchInputPayloadFromFormData(&info); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if merr := models.AuthorizeSMTPEmail(&info); merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
	} else {
		ctl.sendSuccessResp(action, "successfully")
	}
}
