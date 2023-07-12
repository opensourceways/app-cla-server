package controllers

import "github.com/opensourceways/app-cla-server/models"

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

	info.Info = getSingingInfo(info.Info, claInfo.Fields)

	if err := models.SignIndividualCLA(linkID, &info); err != nil {
		if err.IsErrorOf(models.ErrNoLinkOrResigned) {
			ctl.sendFailedResponse(400, errResigned, err, action)
		} else {
			ctl.sendModelErrorAsResp(err, action)
		}

		return
	}

	ctl.sendSuccessResp("sign successfully")
}

// @Title Check
// @Description check whether contributor has signed cla
// @Tags IndividualSigning
// @Accept json
// @Param  link_id  path   string  true  "link id"
// @Param  email    query  string  true  "email of contributor"
// @Success 200 {object} controllers.individualSigned
// @Failure 400 no_link:      there is not link for org
// @Failure 500 system_error: system error
// @router /:link_id [get]
func (ctl *IndividualSigningController) Check() {
	action := "check individual signing"

	v, merr := models.CheckSigning(
		ctl.GetString(":link_id"), ctl.GetString("email"),
	)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
	} else {
		ctl.sendSuccessResp(
			individualSigned{v},
		)
	}
}

type individualSigned struct {
	Signed bool `json:"signed"`
}
