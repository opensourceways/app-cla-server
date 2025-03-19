package controllers

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
)

type CorpEmailDomainController struct {
	baseController
}

func (ctl *CorpEmailDomainController) Prepare() {
	ctl.apiPrepare(PermissionCorpAdmin)
}

// @Title Verify
// @Description send verification code when adding email domain
// @Tags CorpEmailDomain
// @Accept json
// @Param  body  body  controllers.verificationCodeRequest  true  "body for verification code"
// @Success 202 {object} controllers.respData
// @Failure 400 missing_token:      token is missing
// @Failure 401 unknown_token:      token is unknown
// @Failure 402 expired_token:      token is expired
// @Failure 403 unauthorized_token: the permission of token is unauthorized
// @Failure 500 system_error:       system error
// @router /code [post]
func (ctl *CorpEmailDomainController) Verify() {
	action := "corp admin verifies another email domain"
	sendResp := ctl.newFuncForSendingFailedResp(action)

	var req verificationCodeRequest
	if fr := ctl.fetchInputPayload(&req); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if err := req.validate(); err != nil {
		ctl.sendFailedResultAsResp(
			newFailedApiResult(400, errParsingApiBody, err),
			action,
		)
		return
	}

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	orgInfo, merr := models.GetLink(pl.LinkID)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)

		return
	}

	_, cs, merr := models.GetCorpSigning(pl.UserId, pl.SigningId)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)

		return
	}

	code, err := models.VerifyCorpEmailDomain(pl.SigningId, req.Email)
	if err != nil {
		ctl.sendModelErrorAsResp(err, action)
		return
	}

	ctl.sendSuccessResp(action, "successfully")

	sendEmailToIndividual(
		req.Email, &orgInfo,
		"Verification code for adding corporation's another email domain",
		emailtmpl.AddingCorpEmailDomain{
			Corp:       cs.CorporationName,
			Org:        orgInfo.OrgAlias,
			Code:       code,
			ProjectURL: orgInfo.ProjectURL,
		},
	)
}

// @Title Add
// @Description add email domain of corporation
// @Tags CorpEmailDomain
// @Accept json
// @Param  body  body  models.CorpEmailDomainCreateOption  true  "body for email domain"
// @Success 201 {object} controllers.respData
// @Failure 400 missing_token:              token is missing
// @Failure 401 unknown_token:              token is unknown
// @Failure 402 expired_token:              token is expired
// @Failure 403 unauthorized_token:         the permission of token is unauthorized
// @Failure 404 error_parsing_api_body:     fetch payload failed
// @Failure 405 not_an_email:               the email field of payload is not an email
// @Failure 406 expired_verification_code:  the verification code is expired
// @Failure 407 wrong_verification_code:    the verification code is wrong
// @Failure 408 unmatched_email_domain:     the email domain is unmatched
// @Failure 409 no_link_or_unsigned:        no link or corp has not signed
// @Failure 500 system_error:               system error
// @router / [post]
func (ctl *CorpEmailDomainController) Add() {
	action := "corp admin adds email domain"
	sendResp := ctl.newFuncForSendingFailedResp(action)

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	info := &models.CorpEmailDomainCreateOption{}
	if fr := ctl.fetchInputPayload(info); fr != nil {
		sendResp(fr)
		return
	}

	if merr := models.AddCorpEmailDomain(pl.SigningId, info); merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
	} else {
		ctl.sendSuccessResp(action, "successfully")
	}
}

// @Title GetAll
// @Description get all the email domains
// @Tags CorpEmailDomain
// @Accept json
// @Success 200 {object} controllers.respData
// @Failure 400 missing_token:      token is missing
// @Failure 401 unknown_token:      token is unknown
// @Failure 402 expired_token:      token is expired
// @Failure 403 unauthorized_token: the permission of token is unauthorized
// @Failure 500 system_error:       system error
// @router / [get]
func (ctl *CorpEmailDomainController) GetAll() {
	action := "corp admin lists all email domains"

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	r, merr := models.ListCorpEmailDomains(pl.SigningId)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
	} else {
		ctl.sendSuccessResp(action, r)
	}
}
