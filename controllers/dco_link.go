package controllers

import (
	"strings"

	"github.com/opensourceways/app-cla-server/models"
)

type DCOLinkController struct {
	baseController
}

func (ctl *DCOLinkController) Prepare() {
	if strings.HasSuffix(ctl.routerPattern(), "dcos") {
		ctl.apiPrepare("")
	} else {
		ctl.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Create
// @Description create a link(dco application)
// @Tags DCOLink
// @Accept json
// @Param  body  body  models.DCOLinkCreateOption  true  "body for creating link"
// @Success 201 {object} controllers.respData
// @router / [post]
func (ctl *DCOLinkController) Create() {
	action := "community manager creates link(dco application)"
	sendResp := ctl.newFuncForSendingFailedResp(action)

	pl, fr := ctl.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		sendResp(fr)
		return
	}

	input := &models.DCOLinkCreateOption{}
	if fr := ctl.fetchInputPayloadFromFormData(input); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if fr := pl.isOwnerOfOrg(input.Platform, input.OrgID); fr != nil {
		sendResp(fr)
		return
	}

	if merr := models.AddDCOLink(pl.User, input); merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
		return
	}

	ctl.sendSuccessResp(action, "successfully")
}

// @Title Delete
// @Description delete link
// @Tags DCOLink
// @Accept json
// @Param  link_id  path  string  true  "link id"
// @Success 204 {object} controllers.respData
// @router /:link_id [delete]
func (ctl *DCOLinkController) Delete() {
	linkId := ctl.GetString(":link_id")
	action := "community manager delete dco link: " + linkId
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

	if err := models.RemoveDCOLink(linkId); err != nil {
		ctl.sendModelErrorAsResp(err, action)
		return
	}

	ctl.sendSuccessResp(action, "successfully")
}

// @Title List
// @Description list all links
// @Tags DCOLink
// @Accept json
// @Success 200 {object} models.LinkInfo
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 500 system_error:               system error
// @router / [get]
func (ctl *DCOLinkController) List() {
	action := "community manager list dco links"

	pl, fr := ctl.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if r, merr := models.ListDCOLink(pl.Platform, pl.Orgs); merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
	} else {
		ctl.sendSuccessResp(action, r)
	}
}

// @Title GetDCOForSigning
// @Description get signing page info
// @Tags DCOLink
// @Accept json
// @Param  link_id   path  string  true  "link id"
// @Param  apply_to  path  string  true  "apply to"
// @Success 200 {object} models.CLADetail
// @Failure util.ErrNoCLABindingDoc	"org has not been bound any clas"
// @Failure util.ErrNotReadyToSign	"the corp signing is not ready"
// @router /:link_id/dcos [get]
func (ctl *DCOLinkController) GetDCOForSigning() {
	action := "fetch dco signing page info"

	result, err := models.ListDCOs(ctl.GetString(":link_id"))
	if err != nil {
		ctl.sendModelErrorAsResp(err, action)
	} else {
		ctl.sendSuccessResp(action, result)
	}
}