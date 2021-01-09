package controllers

import (
	"fmt"
	"os"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/util"
)

type CLAController struct {
	baseController
}

func (this *CLAController) Prepare() {
	apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg})
}

// @Title Link
// @Description link org and cla
// @Param	body		body 	models.OrgCLA	true		"body for org-repo content"
// @Success 201 {int} models.OrgCLA
// @Failure 403 body is empty
// @router /:link_id/:apply_to [post]
func (this *CLAController) Add() {
	doWhat := "add cla"
	linkID := this.GetString(":link_id")
	applyTo := this.GetString(":apply_to")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}

	input := &models.CLACreateOpt{}
	if fr := this.fetchInputPayload(input); fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}

	if applyTo == dbmodels.ApplyToCorporation {
		data, fr := this.readInputFile(fileNameOfUploadingOrgSignatue)
		if fr != nil {
			this.sendFailedResultAsResp(fr, doWhat)
			return
		}
		input.SetOrgSignature(&data)
	}

	if merr := input.Validate(applyTo, pdf.GetPDFGenerator().LangSupported()); merr != nil {
		this.sendModelErrorAsResp(merr, doWhat)
		return
	}

	orgInfo := pl.orgInfo(linkID)
	filePath := genOrgFileLockPath(orgInfo.Platform, orgInfo.OrgID, orgInfo.RepoID)
	unlock, err := util.Lock(filePath)
	if err != nil {
		this.sendFailedResponse(500, errSystemError, err, doWhat)
		return
	}
	defer unlock()

	if fr := addCLA(linkID, applyTo, input); fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}

	this.sendSuccessResp("add cla successfully")
}

// @Title Delete CLA
// @Description delete cla
// @Param	uid		path 	string	true		"cla id"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:link_id/:apply_to:/:language [delete]
func (this *CLAController) Delete() {
	doWhat := "delete cla"
	linkID := this.GetString(":link_id")
	applyTo := this.GetString(":apply_to")
	claLang := this.GetString(":language")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}

	orgInfo := pl.orgInfo(linkID)
	filePath := genOrgFileLockPath(orgInfo.Platform, orgInfo.OrgID, orgInfo.RepoID)
	unlock, err := util.Lock(filePath)
	if err != nil {
		this.sendFailedResponse(500, errSystemError, err, doWhat)
		return
	}
	defer unlock()

	if r := deleteCLA(linkID, applyTo, claLang); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, doWhat)
		return
	}

	this.sendSuccessResp("delete cla successfully")
}

// @Title Get
// @Description get cla by uid
// @Param	uid		path 	string	true		"The key for cla"
// @Success 200 {object} models.CLA
// @Failure 403 :uid is empty
// @router /:uid [get]
func (this *CLAController) Get() {
	var statusCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	uid := this.GetString(":uid")
	if uid == "" {
		reason = fmt.Errorf("missing cla id")
		statusCode = 400
		return
	}

	cla := models.CLA{ID: uid}

	if err := (&cla).Get(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = cla
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

	path := genCLAFilePath(linkID, applyTo, input.Language)
	if err := input.SaveCLAAtLocal(path); err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}

	path = genOrgSignatureFilePath(linkID, input.Language)
	if err := input.SaveSignatueAtLocal(path); err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}

	if merr := models.DeleteCLAInfo(linkID, applyTo, input.Language); merr != nil {
		return parseModelError(merr)
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
	_, err := models.GetCLAInfoSigned(linkID, claLang, applyTo)
	if err == nil {
		return newFailedApiResult(400, errCLAIsUsed, fmt.Errorf("cla is used"))
	}
	if !err.IsErrorOf(models.ErrNoLinkOrUnsigned) {
		return parseModelError(err)
	}

	if merr := models.DeleteCLA(linkID, applyTo, claLang); merr != nil {
		return parseModelError(merr)
	}

	models.DeleteCLAInfo(linkID, applyTo, claLang)

	if applyTo == dbmodels.ApplyToCorporation {
		path := genOrgSignatureFilePath(linkID, claLang)
		if !util.IsFileNotExist(path) {
			os.Remove(path)
		}
	}
	return nil
}
