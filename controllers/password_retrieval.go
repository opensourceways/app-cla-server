package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
)

type PasswordRetrievalController struct {
	baseController
}

func (this *PasswordRetrievalController) Prepare() {
	this.stopRunIfSignSerivceIsUnabled()

	this.apiPrepare("")
}

// @Title Post
// @Description retrieving the password by sending an email to the user
// @Param 	link_id		path 	string				true		"link id"
// @Param	body		body 	models.PasswordRetrievalKey	true		"body for retrieving password"
// @Success 201 {string}
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 error_parsing_api_body:     parse payload of request failed
// @Failure 402 no_link:                    the link id is not exists
// @Failure 403 missing_email:              missing email in payload
// @Failure 500 system_error:               system error
// @router /:link_id [post]
func (this *PasswordRetrievalController) Post() {
	action := "send an email to retrieve password"
	linkID := this.GetString(":link_id")

	orgInfo, mErr := models.GetOrgOfLink(linkID)
	if mErr != nil {
		this.sendModelErrorAsResp(mErr, action)
		return
	}

	var info models.PasswordRetrievalKey
	if fr := this.fetchInputPayload(&info); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if err := (&info).Validate(); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	b, mErr := info.Create(linkID, config.AppConfig.VerificationCodeExpiry)
	if mErr != nil {
		this.sendModelErrorAsResp(mErr, action)
		return
	}

	key, err := encryptData(b)
	if err != nil {
		this.sendFailedResponse(500, errSystemError, err, action)
		return
	}

	this.sendSuccessResp(action + "successfully")

	sendEmailToIndividual(
		info.Email,
		orgInfo.OrgEmail,
		"[CLA Sign] Retrieving Password of Corporation Manager",
		email.PasswordRetrieval{
			Org:          orgInfo.OrgAlias,
			Timeout:      config.AppConfig.PasswordRetrievalExpiry / 60,
			ResetURL:     config.AppConfig.GenURLToResetPassword(linkID, key),
			RetrievalURL: config.AppConfig.GenURLToRetrievalPassword(linkID),
		},
	)
}

// @Title Reset
// @Description retrieve password of corporation manager by resetting it
// @Param 	body		body 	models.PasswordRetrieval 	true 	"body of retrieving password"
// @Success 201 {string}
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 error_parsing_api_body:     parse payload of request failed
// @Failure 402 missing_pw_retrieval_key:   missing password retrieval key in header
// @Failure 403 invalid_pw_retrieval_key:   invalid password retrieval key
// @Failure 404 expired_verification_code:  the verification code is expired
// @Failure 405 wrong_verification_code:    the verification code is wrong
// @Failure 406 no_link_or_no_manager:      invalid password retrieval key
// @Failure 406 invalid_password:           invalid new password
// @Failure 500 system_error:               system error
// @router /:link_id [patch]
func (this *PasswordRetrievalController) Reset() {
	action := "retrieve password of corporation manager"
	sendResp := this.newFuncForSendingFailedResp(action)

	key := this.apiReqHeader(headerPasswordRetrievalKey)
	if key == "" {
		this.sendFailedResponse(
			400, errMissingPWRetrievalKey,
			fmt.Errorf("missing password retrival key"), action,
		)
		return
	}

	b, err := decryptData(key)
	if err != nil {
		this.sendFailedResponse(
			400, errInvalidPWRetrievalKey,
			fmt.Errorf("invalid password retrival key"), action,
		)
		return
	}

	var param models.PasswordRetrieval
	if fr := this.fetchInputPayload(&param); fr != nil {
		sendResp(fr)
		return
	}

	if mErr := param.Validate(); mErr != nil {
		this.sendModelErrorAsResp(mErr, action)
		return
	}

	if mErr := param.Create(this.GetString(":link_id"), b); mErr != nil {
		sendResp(parseModelError(mErr))
		return
	}

	this.sendSuccessResp(action + "successfully")
}
