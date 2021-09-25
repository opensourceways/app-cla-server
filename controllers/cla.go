package controllers

import (
	"fmt"
	"os"
	"strings"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/util"
)

type CLAController struct {
	baseController
}

func (this *CLAController) Prepare() {
	if isSigningServiceNotStarted() {
		this.StopRun()
	}

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

	if merr := input.Validate(applyTo, pdf.GetPDFGenerator().LangSupported()); merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	unlock, fr := lockOnRepo(pl.orgInfo(linkID))
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	defer unlock()

	if fr := addCLA(linkID, applyTo, input); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	this.sendSuccessResp("add cla successfully")
}

// @Title Delete CLA
// @Description delete cla
// @Param	uid		path 	string	true		"cla id"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:link_id/:apply_to/:language [delete]
func (this *CLAController) Delete() {
	action := "delete cla"
	linkID := this.GetString(":link_id")
	applyTo := this.GetString(":apply_to")
	claLang := this.GetString(":language")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	unlock, fr := lockOnRepo(pl.orgInfo(linkID))
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	defer unlock()

	if r := deleteCLA(linkID, applyTo, claLang); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, action)
		return
	}

	this.sendSuccessResp("delete cla successfully")
}

// @Title Download CLA PDF
// @Description get cla pdf
// @Success 200
// @router /:link_id/:apply_to/:language/:hash [get]
func (this *CLAController) DownloadPDF() {
	path := genCLAFilePath(
		this.GetString(":link_id"),
		this.GetString(":apply_to"),
		this.GetString(":language"),
		this.GetString(":hash"),
	)

	this.downloadFile(path)
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

	clas, merr := models.GetAllCLA(linkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendSuccessResp(clas)
}

func addCLA(linkID, applyTo string, input *models.CLACreateOpt) *failedApiResult {
	hasCLA, merr := models.HasCLA(linkID, applyTo, input.Language)
	if merr != nil {
		return parseModelError(merr)
	}
	if hasCLA {
		return newFailedApiResult(400, errCLAExists, fmt.Errorf("recreate cla"))
	}

	if merr := models.DeleteCLAInfo(linkID, applyTo, input.Language); merr != nil {
		return parseModelError(merr)
	}

	if fr := saveCLAPDF(input, linkID, applyTo); fr != nil {
		return fr
	}

	if merr := input.AddCLAInfo(linkID, applyTo); merr != nil {
		return parseModelError(merr)
	}

	if merr := input.AddCLA(linkID, applyTo); merr != nil {
		return parseModelError(merr)
	}

	return nil
}

func deleteCLA(linkID, applyTo, claLang string) *failedApiResult {
	claInfo, fr := getCLAInfoSigned(linkID, claLang, applyTo)
	if fr != nil {
		return fr
	}
	if claInfo != nil {
		return newFailedApiResult(400, errCLAIsUsed, fmt.Errorf("cla is used"))
	}

	if merr := models.DeleteCLA(linkID, applyTo, claLang); merr != nil {
		return parseModelError(merr)
	}

	models.DeleteCLAInfo(linkID, applyTo, claLang)
	deleteCLAPDF(linkID, applyTo, claInfo)
	return nil
}

func deleteCLAPDF(linkID, applyTo string, claInfo *dbmodels.CLAInfo) *failedApiResult {
	path := genCLAFilePath(linkID, applyTo, claInfo.CLALang, claInfo.CLAHash)
	if !util.IsFileNotExist(path) {
		os.Remove(path)
	}

	key := models.CLAPDFIndex{
		LinkID: linkID,
		Apply:  applyTo,
		Lang:   claInfo.CLALang,
		Hash:   claInfo.CLAHash,
	}
	err := models.DeleteCLAPDF(key)
	return parseModelError(err)
}

func saveCLAPDF(cla *models.CLACreateOpt, linkID, applyTo string) *failedApiResult {
	if cla == nil {
		return nil
	}

	path := genCLAFilePath(linkID, applyTo, cla.Language, cla.GetCLAHash())
	if err := cla.SaveCLAAtLocal(path); err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}

	if err := cla.UploadCLAPDF(linkID, applyTo); err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}

	return nil
}
