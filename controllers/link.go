package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/util"
)

type LinkController struct {
	baseController
}

func (this *LinkController) Prepare() {
	if this.routerPattern() == "/v1/link/:platform/:org_id/:apply_to" {
		if this.apiReqHeader(headerToken) != "" {
			this.apiPrepare(PermissionIndividualSigner)
		}
	} else {
		this.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Link
// @Description link org and cla
// @Param	body		body 	models.LinkCreateOption	true		"body for creating link"
// @Success 201 {string} "create org cla successfully"
// @router / [post]
func (this *LinkController) Link() {
	action := "create link"
	sendResp := this.newFuncForSendingFailedResp(action)

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		sendResp(fr)
		return
	}

	input := &models.LinkCreateOption{}
	if fr := this.fetchInputPayloadFromFormData(input); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if input.CorpCLA != nil {
		data, fr := this.readInputFile(fileNameOfUploadingOrgSignatue)
		if fr != nil {
			sendResp(fr)
			return
		}
		input.CorpCLA.SetOrgSignature(&data)
	}

	if merr := input.Validate(pdf.GetPDFGenerator().LangSupported()); merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	if fr := pl.isOwnerOfOrg(input.OrgID); fr != nil {
		sendResp(fr)
		return
	}

	filePath := genOrgFileLockPath(input.Platform, input.OrgID, input.RepoID)
	if err := util.CreateLockedFile(filePath); err != nil {
		this.sendFailedResponse(500, errSystemError, err, action)
		return
	}

	unlock, err := util.Lock(filePath)
	if err != nil {
		this.sendFailedResponse(500, errSystemError, err, action)
		return
	}
	defer unlock()

	orgRepo := buildOrgRepo(input.Platform, input.OrgID, input.RepoID)
	_, merr := models.GetLinkID(orgRepo)
	if merr == nil {
		this.sendFailedResponse(400, errLinkExists, fmt.Errorf("recreate link"), action)
		return
	}
	if !merr.IsErrorOf(models.ErrNoLink) {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	linkID := genLinkID(orgRepo)
	if fr := this.writeLocalFileOfLink(input, linkID); fr != nil {
		sendResp(fr)
		return
	}

	if fr := this.initializeSigning(input, linkID, orgRepo); fr != nil {
		sendResp(fr)
		return
	}

	if merr := input.Create(linkID, pl.User); merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendResponse("create org cla successfully", 0)
}

func (this *LinkController) writeLocalFileOfLink(input *models.LinkCreateOption, linkID string) *failedApiResult {
	cla := input.CorpCLA
	if cla != nil {
		path := genCLAFilePath(linkID, dbmodels.ApplyToCorporation, cla.Language)
		if err := cla.SaveCLAAtLocal(path); err != nil {
			return newFailedApiResult(500, errSystemError, err)
		}

		path = genOrgSignatureFilePath(linkID, cla.Language)
		if err := cla.SaveSignatueAtLocal(path); err != nil {
			return newFailedApiResult(500, errSystemError, err)
		}
	}

	cla = input.IndividualCLA
	if cla != nil {
		path := genCLAFilePath(linkID, dbmodels.ApplyToIndividual, cla.Language)
		if err := cla.SaveCLAAtLocal(path); err != nil {
			return newFailedApiResult(500, errSystemError, err)
		}
	}

	return nil
}

func (this *LinkController) initializeSigning(input *models.LinkCreateOption, linkID string, orgRepo *dbmodels.OrgRepo) *failedApiResult {
	var info *dbmodels.CLAInfo

	if input.IndividualCLA != nil {
		info = input.IndividualCLA.GenCLAInfo()
	}
	if merr := models.InitializeIndividualSigning(linkID, info); merr != nil {
		return parseModelError(merr)
	}

	orgInfo := dbmodels.OrgInfo{
		OrgRepo:  *orgRepo,
		OrgEmail: input.OrgEmail,
		OrgAlias: input.OrgAlias,
	}
	info = nil
	if input.CorpCLA != nil {
		info = input.CorpCLA.GenCLAInfo()
	}
	if merr := models.InitializeCorpSigning(linkID, &orgInfo, info); merr != nil {
		return parseModelError(merr)
	}

	return nil
}

// @Title Unlink
// @Description unlink cla
// @Param	uid		path 	string	true		"The uid of binding"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:link_id [delete]
func (this *LinkController) Unlink() {
	action := "unlink"
	sendResp := this.newFuncForSendingFailedResp(action)
	linkID := this.GetString(":link_id")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		sendResp(fr)
		return
	}

	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		sendResp(fr)
		return
	}

	if err := models.Unlink(linkID); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp(action + "successfully")
}

// @Title ListOrgs
// @Description get all orgs
// @Success 200 {object} models.OrgInfo
// @router / [get]
func (this *LinkController) ListLinks() {
	action := "list links"

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if len(pl.Orgs) == 0 {
		this.sendSuccessResp(nil)
		return
	}

	orgs := make([]string, 0, len(pl.Orgs))
	for k := range pl.Orgs {
		orgs = append(orgs, k)
	}
	r, merr := models.ListLinks(pl.Platform, orgs)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendSuccessResp(r)
}

// @Title GetCLAForSigning
// @Description get signing page info
// @Param	:platform	path 	string				true		"code platform"
// @Param	:org_id		path 	string				true		"org"
// @Param	:apply_to	path 	string				true		"apply to"
// @Success 201 {int} map
// @Failure util.ErrNoCLABindingDoc	"this org/repo has not been bound any clas"
// @Failure util.ErrNotReadyToSign	"the corp signing is not ready"
// @router /:platform/:org_id/:apply_to [get]
func (this *LinkController) GetCLAForSigning() {
	action := "fetch signing page info"
	applyTo := this.GetString(":apply_to")
	token := this.apiReqHeader(headerToken)

	if !((token == "" && applyTo == dbmodels.ApplyToCorporation) ||
		(token != "" && applyTo == dbmodels.ApplyToIndividual)) {
		this.sendFailedResponse(400, errUnmatchedCLAType, fmt.Errorf("unmatched cla type"), action)
		return
	}

	org, repo := parseOrgAndRepo(this.GetString(":org_id"))
	orgRepo := buildOrgRepo(this.GetString(":platform"), org, repo)

	if linkID, r, err := models.GetCLAByType(orgRepo, applyTo); err != nil {
		this.sendModelErrorAsResp(err, action)
	} else {
		result := struct {
			LinkID string               `json:"link_id"`
			CLAs   []dbmodels.CLADetail `json:"clas"`
		}{
			LinkID: linkID,
			CLAs:   r,
		}
		this.sendSuccessResp(result)
	}
}
