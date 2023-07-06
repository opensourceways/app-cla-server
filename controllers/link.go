package controllers

import (
	"strings"

	"github.com/opensourceways/app-cla-server/models"
)

type LinkController struct {
	baseController
}

func (this *LinkController) Prepare() {
	if strings.HasSuffix(this.routerPattern(), ":apply_to") {
		this.apiPrepare("")
	} else {
		this.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Link
// @Description link org and cla
// @Param	body		body 	models.LinkCreateOption	true		"body for creating link"
// @Success 201 {string} "create org cla successfully"
// @router / [post]
func (this *LinkController) Link() {
	action := "create link"
	sendResp := this.newFuncForSendingFailedResp(action)

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		sendResp(fr)
		return
	}

	input := &models.LinkCreateOption{}
	if fr := this.fetchInputPayloadFromFormData(input); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if fr := pl.isOwnerOfOrg(input.Platform, input.OrgID); fr != nil {
		sendResp(fr)
		return
	}

	if merr := models.AddLink(pl.User, input); merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendResponse("create org cla successfully", 0)
}

// @Title Unlink
// @Description unlink cla
// @Param	uid		path 	string	true		"The uid of binding"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:link_id [delete]
func (this *LinkController) Unlink() {
	action := "unlink"
	sendResp := this.newFuncForSendingFailedResp(action)
	linkID := this.GetString(":link_id")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		sendResp(fr)
		return
	}

	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		sendResp(fr)
		return
	}

	if err := models.RemoveLink(linkID); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp(action + "successfully")
}

// @Title ListLinks
// @Description list all links
// @Success 200 {object} dbmodels.LinkInfo
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 500 system_error:               system error
// @router / [get]
func (this *LinkController) ListLinks() {
	action := "list links"

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if r, merr := models.ListLink(pl.Platform, pl.Orgs); merr != nil {
		this.sendModelErrorAsResp(merr, action)
	} else {
		this.sendSuccessResp(r)
	}
}

// @Title GetCLAForSigning
// @Description get signing page info
// @Param	:link_id	path 	string				true		"link id"
// @Param	:apply_to	path 	string				true		"apply to"
// @Success 201 {int} map
// @Failure util.ErrNoCLABindingDoc	"this org/repo has not been bound any clas"
// @Failure util.ErrNotReadyToSign	"the corp signing is not ready"
// @router /:link_id/:apply_to [get]
func (this *LinkController) GetCLAForSigning() {
	action := "fetch signing page info"

	result, err := models.ListCLAs(
		this.GetString(":link_id"), this.GetString(":apply_to"),
	)
	if err != nil {
		this.sendModelErrorAsResp(err, action)
	} else {
		this.sendSuccessResp(result)
	}
}

// @Title UpdateLinkEmail
// @Description update link email
// @Param  :link_id  path  string  true	 "link id"
// @router /update/:link_id [post]
func (this *LinkController) UpdateLinkEmail() {
	this.sendSuccessResp("unimplemented")
	return
}
