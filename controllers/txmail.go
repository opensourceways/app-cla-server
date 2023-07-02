package controllers

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/txmailimpl"
)

type TXmailController struct {
	baseController
}

func (this *TXmailController) Prepare() {
	this.stopRunIfSignSerivceIsUnabled()

	this.apiPrepare(PermissionOwnerOfOrg)
}

// @Title Code
// @Description send Email authorization verification code
// @router /code [post]
func (this *TXmailController) Code() {
	action := "send Email authorization verification code"

	var info models.EmailAuthorizationReq
	if fr := this.fetchInputPayloadFromFormData(&info); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	code, me := models.CreateCodeForSettingOrgEmail(info.Email)
	if me != nil {
		this.sendModelErrorAsResp(me, action)
		return
	}

	e := emailtmpl.EmailVerification{
		Code: code,
	}
	msg, err := e.GenEmailMsg()
	if err != nil {
		this.sendFailedResponse(500, errSystemError, err, action)

		return
	}

	msg.From = info.Email
	msg.To = []string{info.Email}
	msg.Subject = "CLA Email authorization verification code"

	if err = txmailimpl.TXmailClient().Send(info.Authorize, msg); err != nil {
		this.sendFailedResponse(400, errInvalidEmailAuthCode, err, action)
	} else {
		this.sendSuccessResp("succss")
	}
}

// @Title Authorize
// @Description Email authorization verification
// @router /authorize [post]
func (this *TXmailController) Authorize() {
	action := "Email authorization verification"

	var info models.EmailAuthorization
	if fr := this.fetchInputPayloadFromFormData(&info); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if verr := (&info).Validate(); verr != nil {
		this.sendModelErrorAsResp(verr, action)
		return
	}

	if cerr := models.AddTxmailCredential(&info.EmailAuthorizationReq); cerr != nil {
		this.sendModelErrorAsResp(cerr, action)
	} else {
		this.sendSuccessResp("Email Authorization Success")
	}
}
