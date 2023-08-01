package controllers

import (
	"strings"

	"github.com/opensourceways/app-cla-server/models"
)

type DCOController struct {
	baseController
}

func (ctl *DCOController) Prepare() {
	if ctl.isGetRequest() && strings.HasSuffix(ctl.routerPattern(), "/:link_id/:id") {
		ctl.apiPrepare("")
	} else {
		ctl.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Add
// @Description add dco
// @Tags DCO
// @Accept json
// @Param  body  body  models.DCOCreateOpt  true  "body for adding dco"
// @Success 201 {object} controllers.respData
// @router /:link_id [post]
func (ctl *DCOController) Add() {
	action := "add dco"
	linkID := ctl.GetString(":link_id")

	pl, fr := ctl.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	input := &models.DCOCreateOpt{}
	if fr := ctl.fetchInputPayloadFromFormData(input); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if err := models.AddDCOInstance(linkID, input); err != nil {
		ctl.sendModelErrorAsResp(err, action)
	} else {
		ctl.sendSuccessResp(action, "successfully")
	}
}

// @Title Delete
// @Description delete dco
// @Tags DCO
// @Accept json
// @Param  link_id  path  string  true  "link id"
// @Param  id       path  string  true  "dco id"
// @Success 204 {object} controllers.respData
// @router /:link_id/:id [delete]
func (ctl *DCOController) Delete() {
	action := "delete dco"
	linkID := ctl.GetString(":link_id")
	dcoId := ctl.GetString(":id")

	pl, fr := ctl.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if err := models.RemoveDCOInstance(linkID, dcoId); err != nil {
		ctl.sendModelErrorAsResp(err, action)

		return
	}

	ctl.sendSuccessResp(action, "successfully")
}

// @Title DownloadPDF
// @Description get dco pdf
// @Tags DCO
// @Accept json
// @Param  link_id  path  string  true  "link id"
// @Param  id       path  string  true  "dco id"
// @Success 200
// @router /:link_id/:id [get]
func (ctl *DCOController) DownloadPDF() {
	ctl.downloadFile(models.DCOFile(
		ctl.GetString(":link_id"), ctl.GetString(":id"),
	))
}

// @Title List
// @Description list dcos of link
// @Tags DCO
// @Accept json
// @Param  link_id  path  string  true  "link id"
// @Success 200 {object} models.CLADetail
// @router /:link_id [get]
func (ctl *DCOController) List() {
	action := "list dco"
	linkID := ctl.GetString(":link_id")

	pl, fr := ctl.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if dcos, merr := models.ListDCOInstances(linkID); merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
	} else {
		ctl.sendSuccessResp(action, dcos)
	}
}
