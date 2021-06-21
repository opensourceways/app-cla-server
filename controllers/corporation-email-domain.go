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
// @Failure 400 missing_token:      token is missing
// @Failure 401 unknown_token:      token is unknown
// @Failure 402 expired_token:      token is expired
// @Failure 403 unauthorized_token: the permission of token is unauthorized
// @Failure 500 system_error:       system error
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
