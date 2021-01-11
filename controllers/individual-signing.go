package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
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

	fr = signHelper(
		linkID, claLang, dbmodels.ApplyToIndividual,
		func(claInfo *models.CLAInfo) *failedApiResult {
			if claInfo.CLAHash != this.GetString(":cla_hash") {
				return newFailedApiResult(400, errUnmatchedCLA, fmt.Errorf("invalid cla"))
			}

			info.Info = getSingingInfo(info.Info, claInfo.Fields)

			if err := (&info).Create(linkID, true); err != nil {
				if err.IsErrorOf(models.ErrNoLinkOrResigned) {
					return newFailedApiResult(400, errResigned, err)
				}
				return parseModelError(err)
			}
			return nil
		},
	)
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
	} else {
		this.sendSuccessResp("sign successfully")
	}
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
	action := "check individual signing"
	org, repo := parseOrgAndRepo(this.GetString(":org_repo"))

	linkID, err := models.GetLinkID(buildOrgRepo(this.GetString(":platform"), org, repo))
	if err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	if v, merr := models.IsIndividualSigned(linkID, this.GetString("email")); merr != nil {
		this.sendModelErrorAsResp(merr, action)
	} else {
		this.sendSuccessResp(map[string]bool{"signed": v})
	}
}
