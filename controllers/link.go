package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
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
// @Param	body		body 	models.OrgCLA	true		"body for org-repo content"
// @Success 201 {int} models.OrgCLA
// @Failure 403 body is empty
// @router / [post]
func (this *LinkController) Link() {
	doWhat := "create link"

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	input, err := this.fetchPayloadOfCreatingLink()
	if err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, doWhat)
		return
	}

	if input.CorpCLA != nil {
		data, fr := this.readInputFile(fileNameOfUploadingOrgSignatue)
		if fr != nil {
			this.sendFailedResultAsResp(fr, doWhat)
			return
		}
		input.CorpCLA.SetOrgSignature(&data)
	}

	if ec, err := input.Validate(); err != nil {
		this.sendFailedResponse(400, ec, err, doWhat)
		return
	}

	beego.Info("abc")

	if r := pl.isOwnerOfOrg(input.OrgID); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, doWhat)
		return
	}

	filePath := genOrgFileLockPath(input.Platform, input.OrgID, input.RepoID)
	if err := this.createFileLock(filePath); err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	unlock, err := util.Lock(filePath)
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}
	defer unlock()

	orgRepo := buildOrgRepo(input.Platform, input.OrgID, input.RepoID)
	hasLink, err := models.HasLink(orgRepo)
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}
	if hasLink {
		this.sendFailedResponse(400, util.ErrRecordExists, fmt.Errorf("recreate link"), doWhat)
		return
	}

	linkID := genLinkID(orgRepo)
	if fr := this.writeLocalFileOfLink(input, linkID); fr != nil {
		this.sendFailedResponse(fr.statusCode, fr.errCode, fr.reason, doWhat)
		return
	}

	if fr := this.initializeSigning(input, linkID, orgRepo); fr != nil {
		this.sendFailedResponse(fr.statusCode, fr.errCode, fr.reason, doWhat)
		return
	}

	beego.Info("input.Create")
	if _, err := input.Create(linkID, pl.User); err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	this.sendResponse("create org cla successfully", 0)
}

func (this *LinkController) fetchPayloadOfCreatingLink() (*models.LinkCreateOption, error) {
	input := &models.LinkCreateOption{}
	if err := json.Unmarshal([]byte(this.Ctx.Request.FormValue("data")), input); err != nil {
		return nil, fmt.Errorf("invalid input payload: %s", err.Error())
	}
	return input, nil
}

func (this *LinkController) createFileLock(path string) error {
	// Create a file as a file lock for this org. The reasons is:
	// A file lock needs the file exist first
	if util.IsFileNotExist(path) {
		return util.CreateLockedFile(path)
	}
	return nil
}

func (this *LinkController) writeLocalFileOfLink(input *models.LinkCreateOption, linkID string) *failedResult {
	cla := input.CorpCLA
	if cla != nil {
		path := genCLAFilePath(linkID, dbmodels.ApplyToCorporation, cla.Language)
		if err := cla.SaveCLAAtLocal(path); err != nil {
			return newFailedResult(500, util.ErrSystemError, err)
		}

		path = genOrgSignatureFilePath(linkID, cla.Language)
		if err := cla.SaveSignatueAtLocal(path); err != nil {
			return newFailedResult(500, util.ErrSystemError, err)
		}
	}

	cla = input.IndividualCLA
	if cla != nil {
		path := genCLAFilePath(linkID, dbmodels.ApplyToIndividual, cla.Language)
		if err := cla.SaveCLAAtLocal(path); err != nil {
			return newFailedResult(500, util.ErrSystemError, err)
		}
	}

	return nil
}

func (this *LinkController) initializeSigning(input *models.LinkCreateOption, linkID string, orgRepo *dbmodels.OrgRepo) *failedResult {
	cla := input.CorpCLA
	if cla != nil {
		orgInfo := dbmodels.OrgInfo{
			OrgRepo:  *orgRepo,
			OrgEmail: input.OrgEmail,
			OrgAlias: input.OrgAlias,
		}
		info := cla.GenCLAInfo()

		if err := models.InitializeCorpSigning(linkID, &orgInfo, info); err != nil {
			return newFailedResult(500, util.ErrSystemError, err)
		}
	}

	cla = input.IndividualCLA
	if cla != nil {
		if err := models.InitializeIndividualSigning(linkID, orgRepo, cla.GenCLAInfo()); err != nil {
			return newFailedResult(500, util.ErrSystemError, err)
		}
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
	doWhat := "unlink"
	linkID := this.GetString(":link_id")

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	if r := pl.isOwnerOfLink(linkID); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, doWhat)
		return
	}

	if err := models.Unlink(linkID); err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	this.sendResponse("unbinding successfully", 0)
}

// @Title ListOrgs
// @Description get all orgs
// @Success 200 {object} models.OrgInfo
// @router / [get]
func (this *LinkController) ListLinks() {
	doWhat := "list links"

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	if len(pl.Orgs) == 0 {
		this.sendResponse(nil, 0)
		return
	}

	orgs := make([]string, 0, len(pl.Orgs))
	for k := range pl.Orgs {
		orgs = append(orgs, k)
	}
	r, err := models.ListLinks(pl.Platform, orgs)
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	this.sendResponse(r, 0)
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
	doWhat := "fetch signing page info"
	applyTo := this.GetString(":apply_to")

	token := this.apiReqHeader(headerToken)
	if !((token == "" && applyTo == dbmodels.ApplyToCorporation) ||
		(token != "" && applyTo == dbmodels.ApplyToIndividual)) {
		this.sendFailedResponse(400, util.ErrInvalidParameter, fmt.Errorf("unmatched cla type"), doWhat)
		return
	}

	org, repo := parseOrgAndRepo(this.GetString(":org_id"))
	orgRepo := buildOrgRepo(this.GetString(":platform"), org, repo)

	if r, err := models.GetCLAByType(orgRepo, applyTo); err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
	} else {
		this.sendResponse(r, 0)
	}
}
