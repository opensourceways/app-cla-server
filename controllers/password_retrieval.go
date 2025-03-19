package controllers

import (
	"fmt"
	"net/url"
	"path"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
)

type PasswordRetrievalController struct {
	baseController
}

func (ctl *PasswordRetrievalController) Prepare() {
	ctl.apiPrepare("")
}

// @Title Post
// @Description retrieving the password by sending an email to the user
// @Tags PasswordRetrieval
// @Accept json
// @Param  body     body  models.PasswordRetrievalKey  true  "body for retrieving password"
// @Success 201 {object} controllers.respData
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 error_parsing_api_body:     parse payload of request failed
// @Failure 402 no_link:                    the link id is not exists
// @Failure 403 missing_email:              missing email in payload
// @Failure 500 system_error:               system error
// @router / [post]
func (ctl *PasswordRetrievalController) Post() {
	action := "manager tries to retrieve password"

	var info models.PasswordRetrievalKey
	if fr := ctl.fetchInputPayload(&info); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if err := (&info).Validate(); err != nil {
		ctl.sendModelErrorAsResp(err, action)
		return
	}

	key, mErr := models.GenKeyForPasswordRetrieval(&info)
	if mErr != nil {
		ctl.sendModelErrorAsResp(mErr, action)
		return
	}

	var linkId string
	var orgInfo models.OrgInfo

	if info.IsCommunityManager() {
		orgInfo = models.OrgInfo{
			OrgEmail:         config.CLAEmailAddr,
			OrgEmailPlatform: config.CLAEmailPlatform,
		}
	} else {
		linkId = info.LinkId

		if orgInfo, mErr = models.GetLink(info.LinkId); mErr != nil {
			ctl.sendModelErrorAsResp(mErr, action)
			return
		}
	}

	ctl.sendSuccessResp(action, "successfully")

	sendEmailToIndividual(
		info.Email,
		&orgInfo,
		"[CLA Sign] Retrieving Password",
		emailtmpl.PasswordRetrieval{
			Org:          orgInfo.OrgAlias,
			Timeout:      config.PasswordRetrievalExpiry / 60,
			ResetURL:     genURLToResetPassword(info.LinkId, key),
			RetrievalURL: genURLToRetrievalPassword(linkId),
		},
	)

	ctl.addOperationLog("", action+", link id: "+info.LinkId, 0)
}

// @Title Reset
// @Description retrieve password of corporation manager by resetting it
// @Tags PasswordRetrieval
// @Accept json
// @Param  body  body  models.PasswordRetrieval  true  "body of retrieving password"
// @Success 202 {object} controllers.respData
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 error_parsing_api_body:     parse payload of request failed
// @Failure 402 missing_pw_retrieval_key:   missing password retrieval key in header
// @Failure 403 invalid_pw_retrieval_key:   invalid password retrieval key
// @Failure 404 expired_verification_code:  the verification code is expired
// @Failure 405 wrong_verification_code:    the verification code is wrong
// @Failure 406 no_link_or_no_manager:      invalid password retrieval key
// @Failure 406 invalid_password:           invalid new password
// @Failure 500 system_error:               system error
// @router /:link_id [put]
func (ctl *PasswordRetrievalController) Reset() {
	linkId := ctl.GetString(":link_id")
	action := "manager resets password, link id: " + linkId
	sendResp := ctl.newFuncForSendingFailedResp(action)

	key := ctl.apiReqHeader(headerPasswordRetrievalKey)
	if key == "" {
		ctl.sendFailedResponse(
			400, errMissingPWRetrievalKey,
			fmt.Errorf("missing password retrival key"), action,
		)
		return
	}

	var param models.PasswordRetrieval
	if fr := ctl.fetchInputPayload(&param); fr != nil {
		sendResp(fr)
		return
	}

	mErr := models.ResetPassword(linkId, &param, key)
	if mErr != nil {
		sendResp(parseModelError(mErr))
	} else {
		ctl.sendSuccessResp(action, "successfully")

		ctl.addOperationLog("", action, 0)
	}
}

func genURLToResetPassword(linkId, key string) string {
	v, err := url.Parse(config.PasswordResetURL)
	if err != nil {
		logs.Error(err)

		return ""
	}

	q := v.Query()
	q.Add("key", key)
	q.Add("link_id", linkId)
	v.RawQuery = q.Encode()

	return v.String()
}

func genURLToRetrievalPassword(linkId string) string {
	v, err := url.Parse(config.PasswordRetrievalURL)
	if err != nil {
		logs.Error(err)

		return ""
	}

	v.Path = path.Join(v.Path, linkId)
	return v.String()
}
