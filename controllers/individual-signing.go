package controllers

import (
	"fmt"
	"net/http"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type IndividualSigningController struct {
	baseController
}

func (this *IndividualSigningController) Prepare() {
	if getRequestMethod(&this.Controller) == http.MethodPost {
		// sign as individual
		this.apiPrepare(PermissionIndividualSigner)
	} else {
		// check sign
		this.apiPrepare("")
	}
}

// @Title Post
// @Description sign as individual
// @Param	:org_cla_id	path 	string				true		"org cla id"
// @Param	body		body 	models.IndividualSigning	true		"body for individual signing"
// @Success 201 {int} map
// @Failure util.ErrHasSigned
// @router /:link_id/:cla_lang/:cla_hash [post]
func (this *IndividualSigningController) Post() {
	doWhat := "sign as individual"
	linkID := this.GetString(":link_id")
	claLang := this.GetString(":cla_lang")

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	var info models.IndividualSigning
	if err := this.fetchInputPayload(&info); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, doWhat)
		return
	}
	info.CLALanguage = claLang

	if err := (&info).Validate(pl.Email); err != nil {
		this.sendModelErrorAsResp(err, doWhat)
		return
	}

	claInfo, merr := models.GetCLAInfoSigned(linkID, claLang, dbmodels.ApplyToIndividual)
	if merr != nil {
		this.sendModelErrorAsResp(merr, doWhat)
		return
	}
	if claInfo == nil {
		// no contributor signed for this language. lock to avoid the cla to be changed
		// before writing to the db.

		orgRepo, merr := models.GetOrgOfLink(linkID)
		if merr != nil {
			this.sendModelErrorAsResp(merr, doWhat)
			return
		}

		unlock, err := util.Lock(genOrgFileLockPath(orgRepo.Platform, orgRepo.OrgID, orgRepo.RepoID))
		if err != nil {
			this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
			return
		}
		defer unlock()

		claInfo, merr = models.GetCLAInfoToSign(linkID, claLang, dbmodels.ApplyToIndividual)
		if merr != nil {
			this.sendModelErrorAsResp(merr, doWhat)
			return
		}
	}

	if claInfo.CLAHash != this.GetString(":cla_hash") {
		this.sendFailedResponse(400, errUnmatchedCLA, fmt.Errorf("invalid cla"), doWhat)
		return
	}

	info.Info = getSingingInfo(info.Info, claInfo.Fields)

	if merr := (&info).Create(linkID, true); merr != nil {
		if merr.IsErrorOf(models.ErrNoLinkOrResign) {
			this.sendFailedResponse(400, errHasSigned, merr, doWhat)
		} else {
			this.sendModelErrorAsResp(merr, doWhat)
		}
		return
	}

	this.sendResponse("sign successfully", 0)
}

// @Title Check
// @Description check whether contributor has signed cla
// @Param	platform	path 	string	true		"code platform"
// @Param	org		path 	string	true		"org"
// @Param	repo		path 	string	true		"repo"
// @Param	email		query 	string	true		"email"
// @Success 200
// @router /:platform/:org/:repo [get]
func (this *IndividualSigningController) Check() {
	doWhat := "check individual signing"

	v, err := models.IsIndividualSigned(
		buildOrgRepo(this.GetString(":platform"), this.GetString(":org"), this.GetString(":repo")),
		this.GetString("email"),
	)
	if err != nil {
		this.sendFailedResultAsResp(parseModelError(err), doWhat)
		return
	}

	this.sendResponse(map[string]bool{"signed": v}, 0)
}
