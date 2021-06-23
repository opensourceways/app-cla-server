package controllers

import (
	"github.com/opensourceways/app-cla-server/models"
)

type CorpEmailDomainController struct {
	baseController
}

func (this *CorpEmailDomainController) Prepare() {
	this.apiPrepare(PermissionCorpAdmin)
}

// @Title Post
// @Description add email domain of corporation
// @Param	body		body 	models.CorpEmailDomainCreateOption	true		"body for email domain"
// @Success 201 {int} map
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
func (this *CorpEmailDomainController) Post() {
	action := "add email domain"
	sendResp := this.newFuncForSendingFailedResp(action)

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	info := &models.CorpEmailDomainCreateOption{}
	if fr := this.fetchInputPayload(info); fr != nil {
		sendResp(fr)
		return
	}

	if err := info.Validate(pl.Email); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	if merr := info.Create(pl.LinkID, pl.Email); merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	// Don't save the new email domain to the payload and refesh a new token.
	// Because if it refreshes new token failed, then the later operations
	// such as adding/deleting new manager will fail. Besides, the implementation
	// is too coupled to those operations.

	this.sendSuccessResp(action + " successfully")
}

// @Title GetAll
// @Description get all the email domains
// @Success 200 {string}
// @Failure 400 missing_token:      token is missing
// @Failure 401 unknown_token:      token is unknown
// @Failure 402 expired_token:      token is expired
// @Failure 403 unauthorized_token: the permission of token is unauthorized
// @Failure 500 system_error:       system error
// @router / [get]
func (this *CorpEmailDomainController) GetAll() {
	action := "list all domains"

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	r, merr := models.ListCorpEmailDomain(pl.LinkID, pl.Email)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendSuccessResp(r)
}
