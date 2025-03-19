package controllers

import (
	"strings"

	"github.com/opensourceways/app-cla-server/models"
)

type CLAController struct {
	baseController
}

func (ctl *CLAController) Prepare() {
	if ctl.isGetRequest() && strings.HasSuffix(ctl.routerPattern(), "/:link_id/:id") {
		ctl.apiPrepare("")
	} else {
		ctl.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Add
// @Description add cla
// @Tags CLA
// @Accept json
// @Param  body  body  models.CLACreateOpt  true  "body for adding cla"
// @Success 201 {object} controllers.respData
// @router /:link_id [post]
func (ctl *CLAController) Add() {
	action := "add cla"
	linkID := ctl.GetString(":link_id")

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	input := &models.CLACreateOpt{}
	if fr := ctl.fetchInputPayloadFromFormData(input); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if err := models.AddCLAInstance(pl.UserId, linkID, input); err != nil {
		ctl.sendModelErrorAsResp(err, action)
	} else {
		ctl.sendSuccessResp(action, "successfully")
	}
}

// @Title Delete
// @Description delete cla
// @Tags CLA
// @Accept json
// @Param  link_id  path  string  true  "link id"
// @Param  id       path  string  true  "cla id"
// @Success 204 {object} controllers.respData
// @router /:link_id/:id [delete]
func (ctl *CLAController) Delete() {
	action := "delete cla"
	linkID := ctl.GetString(":link_id")
	claId := ctl.GetString(":id")

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if err := models.RemoveCLAInstance(pl.UserId, linkID, claId); err != nil {
		ctl.sendModelErrorAsResp(err, action)

		return
	}

	ctl.sendSuccessResp(action, "successfully")
}

// @Title DownloadPDF
// @Description get cla pdf
// @Tags CLA
// @Accept json
// @Param  link_id  path  string  true  "link id"
// @Param  id       path  string  true  "cla id"
// @Success 200
// @router /:link_id/:id [get]
func (ctl *CLAController) DownloadPDF() {
	ctl.downloadFile(models.CLAFile(
		ctl.GetString(":link_id"), ctl.GetString(":id"),
	))
}

// @Title List
// @Description list clas of link
// @Tags CLA
// @Accept json
// @Param  link_id  path  string  true  "link id"
// @Success 200 {object} models.CLAOfLink
// @router /:link_id [get]
func (ctl *CLAController) List() {
	action := "list cla"
	linkID := ctl.GetString(":link_id")

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if clas, merr := models.ListCLAInstances(pl.UserId, linkID); merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
	} else {
		ctl.sendSuccessResp(action, clas)
	}
}
