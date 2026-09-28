package controllers

import (
	"os"
	"strings"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/models"
)

type CLAController struct {
	baseController
}

func (ctl *CLAController) Prepare() {
	if ctl.isGetRequest() && !strings.HasSuffix(ctl.routerPattern(), "/:link_id") {
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

// @Title Update
// @Description update cla
// @Tags CLA
// @Accept json
// @Param  link_id  path  string  true  "link id"
// @Param  id       path  string  true  "cla id"
// @Success 202 {object} controllers.respData
// @router /:link_id [put]
func (ctl *CLAController) Update() {
	action := "update cla"
	linkID := ctl.GetString(":link_id")

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	input := &models.CLAUpdateOpt{}
	if fr := ctl.fetchInputPayloadFromFormData(input); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if err := models.UpdateCLAInstance(pl.UserId, linkID, input); err != nil {
		ctl.sendModelErrorAsResp(err, action)

		return
	}

	ctl.sendSuccessResp(action, "successfully")
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

// @Title TemplatePDF
// @Description export corp cla pdf with signing template
// @Tags CLA
// @Accept json
// @Param  link_id  path  string  true  "link id"
// @Param  id       path  string  true  "cla id"
// @Success 200
// @router /:link_id/:id/template [get]
func (ctl *CLAController) TemplatePDF() {
	action := "export cla template pdf"
	linkID := ctl.GetString(":link_id")
	claId := ctl.GetString(":id")

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	file, err := models.CorpCLATemplatePDF(pl.UserId, linkID, claId)
	if err != nil {
		ctl.sendModelErrorAsResp(err, action)
		return
	}

	ctl.downloadFile(file)

	if e := os.Remove(file); e != nil {
		logs.Error("remove template pdf failed, err: %s", e.Error())
	}
}

// @Title DiffPreview
// @Description preview diff before importing new cla
// @Tags CLA
// @Accept json
// @Param  link_id  path  string  true  "link id"
// @Param  body    body  models.CLADiffPreviewOpt  true  "body for diff preview"
// @Success 200 {object} controllers.respData
// @router /:link_id/diff-preview [post]
func (ctl *CLAController) DiffPreview() {
	action := "diff preview cla"
	linkID := ctl.GetString(":link_id")

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	input := &models.CLADiffPreviewOpt{}
	if fr := ctl.fetchInputPayloadFromFormData(input); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	result, err := models.DiffPreviewCLA(pl.UserId, linkID, input)
	if err != nil {
		ctl.sendModelErrorAsResp(err, action)
		return
	}

	ctl.sendSuccessResp(action, result)
}

// @Title DownloadDiffPDF
// @Description get diff pdf
// @Tags CLA
// @Accept json
// @Param  link_id  path  string  true  "link id"
// @Param  email    path  string  true  "email"
// @Success 200
// @router /individual/diff/:link_id/:email [get]
func (ctl *CLAController) DownloadDiffPDF() {
	action := "action download diff pdf"

	file, err := models.FindDiffCLAFile(ctl.GetString(":link_id"), ctl.GetString(":email"))
	if err != nil {
		ctl.sendModelErrorAsResp(err, action)

		return
	}

	ctl.downloadFile(file)
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
