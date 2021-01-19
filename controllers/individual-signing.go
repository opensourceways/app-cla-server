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
// @Description sign individual cla
// @Param	:link_id	path 	string				true		"link id"
// @Param	:cla_lang	path 	string				true		"cla language"
// @Param	:cla_hash	path 	string				true		"the hash of cla content"
// @Param	body		body 	dbmodels.IndividualSigningInfo	true		"body for individual signing"
// @Success 201 {string} "sign successfully"
// @Failure 400 error_parsing_api_body: parse input paraemter failed
// @Failure 401 unmatched_email: 	the email is not same as the one which signer sets on the code platform
// @Failure 402 unmatched_user_id: 	the user id is not same as the one which was fetched from code platform
// @Failure 403 unmatched_cla:		the cla hash is not equal to the one of backend server
// @Failure 404 resigned: 		the signer has signed the cla
// @Failure 500 system_error: 		system error
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

	if err := (&info).Validate(pl.User, pl.Email); err != nil {
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
// @Param	org_repo	path 	string	true		"org:repo"
// @Param	email		query 	string	true		"email of contributor"
// @Success 200 {object} map
// @Failure 400 no_link: 	there is not link for this org and repo
// @Failure 500 system_error: 	system error
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
