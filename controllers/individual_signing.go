package controllers

import (
	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/models"
)

type IndividualSigningController struct {
	baseController
}

func (ctl *IndividualSigningController) Prepare() {
	ctl.apiPrepare("")
}

// @Title SendVerificationCode
// @Description send verification code when signing
// @Tags IndividualSigning
// @Accept json
// @Param  link_id  path  string                               true  "link id"
// @Param  body     body  controllers.verificationCodeRequest  true  "body for verification code"
// @Success 201 {object} controllers.respData
// @router /:link_id/code [post]
func (ctl *IndividualSigningController) SendVerificationCode() {
	linkId := ctl.GetString(":link_id")

	ctl.sendVerificationCodeWhenSigning(
		linkId,
		func(email string) (string, models.IModelError) {
			return models.VCOfIndividualSigning(linkId, email)
		},
	)
}

// @Title Sign
// @Description sign individual cla
// @Tags IndividualSigning
// @Accept json
// @Param  link_id  path   string                    true  "link id"
// @Param  body     body   models.IndividualSigning  true  "body for individual signing"
// @Success 201 {object} controllers.respData
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 405 error_parsing_api_body:     parse payload of request failed
// @Failure 406 unmatched_email:            the email is not same as the one which signer sets on the code platform
// @Failure 407 unmatched_user_id:          the user id is not same as the one which was fetched from code platform
// @Failure 408 unmatched_cla:              the cla hash is not equal to the one of backend server
// @Failure 409 resigned:                   the signer has signed the cla
// @Failure 410 no_link:                    the link id is not exists
// @Failure 411 go_to_sign_employee_cla:    should sign employee cla instead
// @Failure 500 system_error:               system error
// @router /:link_id/ [post]
func (ctl *IndividualSigningController) Sign() {
	action := "sign individual cla"
	linkID := ctl.GetString(":link_id")

	var info models.IndividualSigning
	if fr := ctl.fetchInputPayload(&info); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	_, claInfo, merr := models.GetLinkCLA(linkID, info.CLAId)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
		return
	}

	if err := models.SignIndividualCLA(linkID, &info, claInfo.Fields); err != nil {
		if err.IsErrorOf(models.ErrNoLinkOrResigned) {
			ctl.sendFailedResponse(400, errResigned, err, action)
		} else {
			ctl.sendModelErrorAsResp(err, action)
		}

		return
	}

	ctl.sendSuccessResp(action, "successfully")
}

// @Title Agree
// @Description agree individual cla
// @Tags IndividualSigning
// @Accept json
// @Param  link_id  path   string                    true  "link id"
// @Param  body     body   models.IndividualSigning  true  "body for individual signing"
// @Success 201 {object} controllers.respData
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 405 error_parsing_api_body:     parse payload of request failed
// @Failure 406 unmatched_email:            the email is not same as the one which signer sets on the code platform
// @Failure 407 unmatched_user_id:          the user id is not same as the one which was fetched from code platform
// @Failure 410 no_link:                    the link id is not exists
// @Failure 500 system_error:               system error
// @router /:link_id/ [put]
func (ctl *IndividualSigningController) Agree() {
	action := "agree individual cla"
	linkID := ctl.GetString(":link_id")

	var info models.IndividualSigning
	if fr := ctl.fetchInputPayload(&info); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	_, _, err := models.GetLinkCLA(linkID, info.CLAId)
	if err != nil {
		ctl.sendModelErrorAsResp(err, action)
		return
	}

	if err = models.AgreeIndividualCLA(linkID, &info); err != nil {
		ctl.sendModelErrorAsResp(err, action)
	} else {
		ctl.sendSuccessResp(action, "successfully")
	}
}

// @Title ConfirmByToken
// @Description confirm the individual cla update with a one-time token from the notification email, no verification code needed
// @Tags IndividualSigning
// @Accept json
// @Param  token  path  string  true  "one-time cla confirm token"
// @Success 200 {object} controllers.respData
// @Failure 400 invalid_cla_confirm_token:  the confirm link is invalid, expired or already used
// @Failure 400 cla_is_latest:              the cla change has already been confirmed
// @Failure 500 system_error:               system error
// @router /cla-confirm/:token [post]
func (ctl *IndividualSigningController) ConfirmByToken() {
	action := "confirm individual cla by token"
	token := ctl.GetString(":token")

	v, merr := models.ConfirmIndividualCLAByToken(token)
	if merr != nil {
		// [audit] 记录一键确认失败（token 无效/过期/已用），只打 token 前缀用于与生成日志关联。
		logs.Info("[audit] cla_confirm_by_token: result=failed, token_prefix=%s, ip=%s",
			tokenPrefix(token), ctl.Ctx.Input.IP())

		ctl.sendModelErrorAsResp(merr, action)

		return
	}

	// [audit] 记录一键确认成功（4W + CLA 上下文）；email 脱敏，不落完整 token。
	logs.Info("[audit] cla_confirm_by_token: result=success, link_id=%s, recipient=%s, ip=%s, new_cla_id=%s",
		v.LinkId, v.EmailMasked, ctl.Ctx.Input.IP(), v.ClaId)

	ctl.sendSuccessResp(action, v)
}

// tokenPrefix returns the first 8 characters of the token for audit logs,
// never the full value.
func tokenPrefix(token string) string {
	if len(token) <= 8 {
		return token
	}

	return token[:8]
}

// @Title Check
// @Description check whether contributor has signed cla
// @Tags IndividualSigning
// @Accept json
// @Param  link_id  path   string  true  "link id"
// @Param  email    query  string  true  "email of contributor"
// @Param  debug    query  bool    false "debug mode to show internal status information"
// @Success 200 {object} controllers.individualSigned
// @Failure 400 no_link:      there is not link for org
// @Failure 500 system_error: system error
// @router /:link_id [get]
func (ctl *IndividualSigningController) Check() {
	action := "check individual signing"
	debug, _ := ctl.GetBool("debug")

	v, merr := models.CheckSigning(
		ctl.GetString(":link_id"), ctl.GetString("email"),
	)

	// signed + version_matched 覆盖全部状态；调试信息仅在 debug 模式下返回
	if !debug {
		v.DebugInfo = nil
	}

	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
	} else {
		ctl.sendSuccessResp(action, v)
	}
}
