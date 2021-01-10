package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type IndividualSigningController struct {
	baseController
}

func (this *IndividualSigningController) Prepare() {
	// sign as individual
	if this.isPostRequest() {
		this.apiPrepare(PermissionIndividualSigner)
	} else {
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
	action := "sign individual cla"
	linkID := this.GetString(":link_id")
	claLang := this.GetString(":cla_lang")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	var info models.IndividualSigning
	if fr := this.fetchInputPayload(&info); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	info.CLALanguage = claLang

	if err := (&info).Validate(pl.Email); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	claInfo, fr := getCLAInfoSigned(linkID, claLang, dbmodels.ApplyToIndividual)
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if claInfo == nil {
		// no contributor signed for this language. lock to avoid the cla to be changed
		// before writing to the db.

		orgRepo, merr := models.GetOrgOfLink(linkID)
		if merr != nil {
			this.sendModelErrorAsResp(merr, action)
			return
		}

		unlock, err := util.Lock(genOrgFileLockPath(orgRepo.Platform, orgRepo.OrgID, orgRepo.RepoID))
		if err != nil {
			this.sendFailedResponse(500, util.ErrSystemError, err, action)
			return
		}
		defer unlock()

		claInfo, merr = models.GetCLAInfoToSign(linkID, claLang, dbmodels.ApplyToIndividual)
		if merr != nil {
			this.sendModelErrorAsResp(merr, action)
			return
		}
		if claInfo == nil {
			this.sendFailedResponse(500, errSystemError, fmt.Errorf("no cla info, impossible"), action)
			return
		}
	}

	if claInfo.CLAHash != this.GetString(":cla_hash") {
		this.sendFailedResponse(400, errUnmatchedCLA, fmt.Errorf("invalid cla"), action)
		return
	}

	info.Info = getSingingInfo(info.Info, claInfo.Fields)

	if err := (&info).Create(linkID, true); err != nil {
		if err.IsErrorOf(models.ErrNoLinkOrResigned) {
			this.sendFailedResponse(400, errResigned, err, action)
		} else {
			this.sendModelErrorAsResp(err, action)
		}
		return
	}

	this.sendSuccessResp("sign successfully")
}

// @Title Check
// @Description check whether contributor has signed cla
// @Param	platform	path 	string	true		"code platform"
// @Param	org		path 	string	true		"org"
// @Param	repo		path 	string	true		"repo"
// @Param	email		query 	string	true		"email"
// @Success 200
// @router /:platform/:org_repo [get]
func (this *IndividualSigningController) Check() {
	sendResp := this.newFuncForSendingFailedResp("check individual signing")
	org, repo := parseOrgAndRepo(this.GetString(":org_repo"))
	emailOfSigner := this.GetString("email")

	linkID, fr := getLinkID(
		this.GetString(":platform"), org, repo, dbmodels.ApplyToIndividual,
	)
	if fr != nil {
		sendResp(fr)
		return
	}

	if v, merr := models.IsIndividualSigned(linkID, emailOfSigner); merr != nil {
		sendResp(parseModelError(merr))
	} else {
		this.sendSuccessResp(map[string]bool{"signed": v})
	}
}
