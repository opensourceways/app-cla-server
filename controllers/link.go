package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/conf"
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

	var input models.LinkCreateOption
	if err := this.fetchInputPayload(&input); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, doWhat)
		return
	}

	if ec, err := input.Validate(); err != nil {
		this.sendFailedResponse(400, ec, err, doWhat)
		return
	}

	if r := isOwnerOfOrg(pl, input.OrgID); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, doWhat)
		return
	}

	// Create a file as a file lock for this org. The reasons are:
	// 1. A file lock needs the file exist first
	// 2. It is safe to create the file here, evet if creating a org's cla concurrently.
	//    Because it doesn't care the content of locked file
	path := util.LockedFilePath(conf.AppConfig.PDFOrgSignatureDir, input.Platform, input.OrgID, input.RepoID)
	if util.IsFileNotExist(path) {
		if err := util.NewFileLock(path).CreateLockedFile(); err != nil {
			this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
			return
		}
	}

	if _, err := input.Create(pl.User); err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	this.sendResponse("create org cla successfully", 0)
}

// @Title Unlink
// @Description unlink cla
// @Param	uid		path 	string	true		"The uid of binding"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:org_id [delete]
func (this *LinkController) Unlink() {
	doWhat := "unlink"

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	org, repo := parseOrgAndRepo(this.GetString(":org_id"))
	if r := isOwnerOfOrg(pl, org); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, doWhat)
		return
	}

	orgRepo := buildOrgRepo(pl.Platform, org, repo)
	if err := models.Unlink(&orgRepo); err != nil {
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

	r, err := models.ListLinks(pl.Platform, pl.orgs())
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

	if r, err := models.GetCLAByType(&orgRepo, applyTo); err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
	} else {
		this.sendResponse(r, 0)
	}
}
