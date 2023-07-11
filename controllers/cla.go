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

// @Title Link
// @Description link org and cla
// @Param	body		body 	models.OrgCLA	true		"body for org-repo content"
// @Success 201 {int} models.OrgCLA
// @Failure 403 body is empty
// @router /:link_id/:apply_to [post]
func (ctl *CLAController) Add() {
	action := "add cla"
	linkID := ctl.GetString(":link_id")
	applyTo := ctl.GetString(":apply_to")

	pl, fr := ctl.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	input := &models.CLACreateOpt{}
	if fr := ctl.fetchInputPayloadFromFormData(input); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if err := models.AddCLAInstance(linkID, input, applyTo); err != nil {
		ctl.sendModelErrorAsResp(err, action)

		return
	}

	ctl.sendSuccessResp("add cla successfully")
}

// @Title Delete CLA
// @Description delete cla
// @Param	uid		path 	string	true		"cla id"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:link_id/:id [delete]
func (ctl *CLAController) Delete() {
	action := "delete cla"
	linkID := ctl.GetString(":link_id")
	claId := ctl.GetString(":id")

	pl, fr := ctl.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if err := models.RemoveCLAInstance(linkID, claId); err != nil {
		ctl.sendModelErrorAsResp(err, action)

		return
	}

	ctl.sendSuccessResp("delete cla successfully")
}

// @Title Download CLA PDF
// @Description get cla pdf
// @Success 200
// @router /:link_id/:id [get]
func (ctl *CLAController) DownloadPDF() {
	ctl.downloadFile(models.CLAFile(
		ctl.GetString(":link_id"), ctl.GetString(":id"),
	))
}

// @Title List
// @Description list clas of link
// @Param	link_id		path 	string	true		"link id"
// @Success 200 {string} delete success!
// @Failure 403 uid is empty
// @router /:link_id [get]
func (ctl *CLAController) List() {
	action := "list cla"
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

	if clas, merr := models.ListCLAInstances(linkID); merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
	} else {
		ctl.sendSuccessResp(clas)
	}
}
