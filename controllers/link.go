package controllers

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/util"
)

type LinkController struct {
	baseController
}

func (this *LinkController) Prepare() {
	if !strings.HasSuffix(this.routerPattern(), ":apply_to") {
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
	if fr := saveCLAPDF(input.CorpCLA, linkID, dbmodels.ApplyToCorporation); fr != nil {
		sendResp(fr)
		return
	}

	if fr := saveCLAPDF(input.IndividualCLA, linkID, dbmodels.ApplyToIndividual); fr != nil {
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
	if input.CorpCLA != nil {
		info = input.CorpCLA.GenCLAInfo()
	} else {
		info = nil
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

// @Title ListLinks
// @Description list all links
// @Success 200 {object} dbmodels.LinkInfo
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 500 system_error:               system error
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
// @Param	:link_id	path 	string				true		"link id"
// @Param	:apply_to	path 	string				true		"apply to"
// @Success 201 {int} map
// @Failure util.ErrNoCLABindingDoc	"this org/repo has not been bound any clas"
// @Failure util.ErrNotReadyToSign	"the corp signing is not ready"
// @router /:link_id/:apply_to [get]
func (this *LinkController) GetCLAForSigning() {
	action := "fetch signing page info"
	applyTo := this.GetString(":apply_to")

	if applyTo != dbmodels.ApplyToCorporation && applyTo != dbmodels.ApplyToIndividual {
		this.sendFailedResponse(400, errUnmatchedCLAType, fmt.Errorf("unmatched cla type"), action)
		return
	}

	linkID := this.GetString(":link_id")
	if result, err := models.GetCLAByType(linkID, applyTo); err != nil {
		this.sendModelErrorAsResp(err, action)
	} else {
		this.sendSuccessResp(result)
	}
}

func LoadLinks() error {
	links, err := models.GetAllLinks()
	if err != nil {
		return err
	}

	f := func(linkID, apply string, cla *dbmodels.CLADetail) error {
		path := genCLAFilePath(linkID, apply, cla.Language, cla.CLAHash)
		s, _ := util.Md5sumOfFile(path)
		if s == cla.CLAHash {
			return nil
		}

		index := models.CLAPDFIndex{
			LinkID: linkID,
			Apply:  apply,
			Lang:   cla.Language,
			Hash:   cla.CLAHash,
		}
		pdf, err := models.DownloadCLAPDF(index)
		if err != nil {
			return fmt.Errorf("download cla pdf failed, err: %v, index: %v", err, index)
		}

		opt := &models.CLACreateOpt{}
		opt.SetCLAContent(pdf)
		return opt.SaveCLAAtLocal(path)
	}

	for i := range links {
		link := &links[i]

		orgRepo := &link.OrgInfo
		filePath := genOrgFileLockPath(orgRepo.Platform, orgRepo.OrgID, orgRepo.RepoID)
		if err := util.CreateLockedFile(filePath); err != nil {
			return err
		}

		linkID := link.LinkID

		info, err := models.GetAllCLA(linkID)
		if err != nil {
			return err
		}

		for j := range info.CorpCLAs {
			if err := f(linkID, dbmodels.ApplyToCorporation, &info.CorpCLAs[j]); err != nil {
				return err
			}
		}

		for j := range info.IndividualCLAs {
			if err := f(linkID, dbmodels.ApplyToIndividual, &info.IndividualCLAs[j]); err != nil {
				return err
			}
		}
	}

	return nil
}
