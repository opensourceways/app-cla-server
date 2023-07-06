package controllers

import (
	"strings"

	"github.com/opensourceways/app-cla-server/models"
)

type CLAController struct {
	baseController
}

func (this *CLAController) Prepare() {
	if strings.HasSuffix(this.routerPattern(), "/:hash") {
		this.apiPrepare("")
	} else {
		this.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Link
// @Description link org and cla
// @Param	body		body 	models.OrgCLA	true		"body for org-repo content"
// @Success 201 {int} models.OrgCLA
// @Failure 403 body is empty
// @router /:link_id/:apply_to [post]
func (this *CLAController) Add() {
	action := "add cla"
	linkID := this.GetString(":link_id")
	applyTo := this.GetString(":apply_to")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	input := &models.CLACreateOpt{}
	if fr := this.fetchInputPayloadFromFormData(input); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if err := models.AddCLAInstance(linkID, input, applyTo); err != nil {
		this.sendModelErrorAsResp(err, action)

		return
	}

	this.sendSuccessResp("add cla successfully")
}

// @Title Delete CLA
// @Description delete cla
// @Param	uid		path 	string	true		"cla id"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:link_id/:id [delete]
func (this *CLAController) Delete() {
	action := "delete cla"
	linkID := this.GetString(":link_id")
	claId := this.GetString(":id")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if err := models.RemoveCLAInstance(linkID, claId); err != nil {
		this.sendModelErrorAsResp(err, action)

		return
	}

	this.sendSuccessResp("delete cla successfully")
}

// @Title Download CLA PDF
// @Description get cla pdf
// @Success 200
// @router /:link_id/:id [get]
func (this *CLAController) DownloadPDF() {
	this.downloadFile(models.CLAFile(
		this.GetString(":link_id"), this.GetString(":id"),
	))
}

// @Title List
// @Description list clas of link
// @Param	link_id		path 	string	true		"link id"
// @Success 200 {string} delete success!
// @Failure 403 uid is empty
// @router /:link_id [get]
func (this *CLAController) List() {
	action := "list cla"
	linkID := this.GetString(":link_id")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if clas, merr := models.ListCLAInstances(linkID); merr != nil {
		this.sendModelErrorAsResp(merr, action)
	} else {
		this.sendSuccessResp(clas)
	}
}
