package controllers

import (
	"fmt"
	"os"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type CLAController struct {
	baseController
}

func (this *CLAController) Prepare() {
	this.apiPrepare(PermissionOwnerOfOrg)
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

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}
	if r := pl.isOwnerOfLink(linkID); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, doWhat)
		return
	}

	input := &models.CLACreateOption{}
	if err := this.fetchInputPayload(input); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, doWhat)
		return
	}

	if applyTo == dbmodels.ApplyToCorporation {
		data, r := this.readInputFile(fileNameOfUploadingOrgSignatue)
		if r != nil {
			this.sendFailedResponse(r.statusCode, r.errCode, r.reason, doWhat)
			return
		}
		input.SetOrgSignature(&data)
	}

	if merr := input.Validate(applyTo); merr != nil {
		this.sendFailedResultAsResp(parseModelError(merr), doWhat)
		return
	}

	if r := addCLA(linkID, applyTo, input, pl); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, doWhat)
		return
	}

	this.sendResponse("add cla successfully", 0)
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

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}
	if err := pl.isOwnerOfLink(linkID); err != nil {
		this.sendFailedResponse(err.statusCode, err.errCode, err.reason, doWhat)
		return
	}

	if r := deleteCLA(linkID, applyTo, claLang, pl); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, doWhat)
		return
	}

	this.sendResponse("delete cla successfully", 0)
}

// @Title List
// @Description list clas of link
// @Param	link_id		path 	string	true		"link id"
// @Success 200 {string} delete success!
// @Failure 403 uid is empty
// @router /:link_id [get]
func (this *CLAController) List() {
	doWhat := "delete cla"
	linkID := this.GetString(":link_id")

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}

	clas, merr := models.GetAllCLA(linkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, doWhat)
		return
	}

	this.sendResponse(clas, 0)
}

func addCLA(linkID, applyTo string, input *models.CLACreateOption, pl *acForCodePlatformPayload) *failedResult {
	orgInfo := pl.orgInfo(linkID)
	filePath := genOrgFileLockPath(orgInfo.Platform, orgInfo.OrgID, orgInfo.RepoID)
	unlock, err := util.Lock(filePath)
	if err != nil {
		return newFailedResult(500, util.ErrSystemError, err)
	}
	defer unlock()

	hasCLA, merr := models.HasCLA(linkID, applyTo, input.Language)
	if merr != nil {
		return parseModelError(merr)
	}
	if hasCLA {
		return newFailedResult(400, errCLAExists, fmt.Errorf("recreate cla"))
	}

	path := genCLAFilePath(linkID, applyTo, input.Language)
	if err := input.SaveCLAAtLocal(path); err != nil {
		return newFailedResult(0, "", err)
	}

	path = genOrgSignatureFilePath(linkID, input.Language)
	if err := input.SaveSignatueAtLocal(path); err != nil {
		return newFailedResult(0, "", err)
	}

	if merr := input.AddCLAInfo(linkID, applyTo); merr != nil {
		return parseModelError(merr)
	}

	if merr := input.AddCLA(linkID, applyTo); merr != nil {
		return parseModelError(merr)
	}

	return nil
}

func deleteCLA(linkID, applyTo, claLang string, pl *acForCodePlatformPayload) *failedResult {
	orgInfo := pl.orgInfo(linkID)
	filePath := genOrgFileLockPath(orgInfo.Platform, orgInfo.OrgID, orgInfo.RepoID)
	unlock, err := util.Lock(filePath)
	if err != nil {
		return newFailedResult(500, util.ErrSystemError, err)
	}
	defer unlock()

	claInfo, err := models.GetCLAInfoSigned(linkID, claLang, applyTo)
	if err != nil {
		return newFailedResult(0, "", err)
	}
	if claInfo != nil {
		return newFailedResult(400, util.ErrCLAIsUsed, fmt.Errorf("cla is used"))
	}

	if merr := models.DeleteCLA(linkID, applyTo, claLang); merr != nil {
		if merr.IsErrorOf(models.ErrNoLink) {
			return newFailedResult(500, errSystemError, fmt.Errorf("impossible"))
		}
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
