package controllers

import (
	"strings"

	"github.com/opensourceways/app-cla-server/models"
)

type LinkController struct {
	baseController
}

func (ctl *LinkController) Prepare() {
	if strings.HasSuffix(ctl.routerPattern(), ":apply_to") {
		ctl.apiPrepare("")
	} else {
		ctl.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Link
// @Description create a link(cla application)
// @Tags Link
// @Accept json
// @Param  body  body  models.LinkCreateOption  true  "body for creating link"
// @Success 201 {object} controllers.respData
// @router / [post]
func (ctl *LinkController) Link() {
	action := "community manager creates link(cla application)"
	sendResp := ctl.newFuncForSendingFailedResp(action)

	pl, fr := ctl.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		sendResp(fr)
		return
	}

	input := &models.LinkCreateOption{}
	if fr := ctl.fetchInputPayloadFromFormData(input); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if fr := pl.isOwnerOfOrg(input.Platform, input.OrgID); fr != nil {
		sendResp(fr)
		return
	}

	if merr := models.AddLink(pl.User, input); merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
		return
	}

	ctl.sendSuccessResp(action, "successfully")
}

// @Title Delete
// @Description delete link
// @Tags Link
// @Accept json
// @Param  link_id  path  string  true  "link id"
// @Success 204 {object} controllers.respData
// @router /:link_id [delete]
func (ctl *LinkController) Delete() {
	linkId := ctl.GetString(":link_id")
	action := "community manager delete link: " + linkId
	sendResp := ctl.newFuncForSendingFailedResp(action)

	pl, fr := ctl.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		sendResp(fr)
		return
	}

	if fr := pl.isOwnerOfLink(linkId); fr != nil {
		sendResp(fr)
		return
	}

	if err := models.RemoveLink(linkId); err != nil {
		ctl.sendModelErrorAsResp(err, action)
		return
	}

	ctl.sendSuccessResp(action, "successfully")
}

// @Title ListLinks
// @Description list all links
// @Tags Link
// @Accept json
// @Success 200 {object} models.LinkInfo
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 500 system_error:               system error
// @router / [get]
func (ctl *LinkController) ListLinks() {
	action := "community manager list links"

	pl, fr := ctl.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if r, merr := models.ListLink(pl.Platform, pl.Orgs); merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
	} else {
		ctl.sendSuccessResp(action, r)
	}
}

// @Title GetCLAForSigning
// @Description get signing page info
// @Tags Link
// @Accept json
// @Param  link_id   path  string  true  "link id"
// @Param  apply_to  path  string  true  "apply to"
// @Success 200 {object} models.CLADetail
// @Failure util.ErrNoCLABindingDoc	"org has not been bound any clas"
// @Failure util.ErrNotReadyToSign	"the corp signing is not ready"
// @router /:link_id/:apply_to [get]
func (ctl *LinkController) GetCLAForSigning() {
	action := "fetch signing page info"

	result, err := models.ListCLAs(
		ctl.GetString(":link_id"), ctl.GetString(":apply_to"),
	)
	if err != nil {
		ctl.sendModelErrorAsResp(err, action)
	} else {
		ctl.sendSuccessResp(action, result)
	}
}
