package controllers

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
)

type IndividualSigningController struct {
	baseController
}

func (this *IndividualSigningController) Prepare() {
	this.stopRunIfSignSerivceIsUnabled()

	// sign as individual
	if this.isPostRequest() {
		this.apiPrepare("")
	} else {
		if strings.HasSuffix(this.routerPattern(), "/:platform/:org_repo") {
			this.apiPrepare("")
		} else {
			this.apiPrepare(PermissionOwnerOfOrg)
		}
	}
}

// @Title Post
// @Description sign individual cla
// @Param	:link_id	path 	string				true		"link id"
// @Param	:cla_lang	path 	string				true		"cla language"
// @Param	:cla_hash	path 	string				true		"the hash of cla content"
// @Param	body		body 	dbmodels.IndividualSigningInfo	true		"body for individual signing"
// @Success 201 {string} "sign successfully"
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 405 error_parsing_api_body:     parse payload of request failed
// @Failure 406 unmatched_email:            the email is not same as the one which signer sets on the code platform
// @Failure 407 unmatched_user_id:          the user id is not same as the one which was fetched from code platform
// @Failure 408 unmatched_cla:              the cla hash is not equal to the one of backend server
// @Failure 409 resigned:                   the signer has signed the cla
// @Failure 410 no_link:                    the link id is not exists
// @Failure 411 go_to_sign_employee_cla:    should sign employee cla instead
// @Failure 500 system_error:               system error
// @router /:link_id/:cla_lang/:cla_hash [post]
func (this *IndividualSigningController) Post() {
	action := "sign individual cla"
	linkID := this.GetString(":link_id")
	claLang := this.GetString(":cla_lang")

	var info models.IndividualSigning
	if fr := this.fetchInputPayload(&info); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	info.CLALanguage = claLang

	if err := (&info).Validate(linkID); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	b, merr := models.IsCorpSigned(linkID, info.Email)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}
	if b {
		this.sendFailedResponse(400, errGoToSignEmployeeCLA, fmt.Errorf("sign employee cla instead"), action)
		return
	}

	fr := signHelper(
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
// @Param	:link_id	path 	string	true		"link id"
// @Param	email		query 	string	true		"email of contributor"
// @Success 200 {object} map
// @Failure 500 system_error: system error
// @router /:link_id [get]
func (this *IndividualSigningController) Check() {
	action := "check individual signing"

	v, merr := models.IsIndividualSigned(
		this.GetString(":link_id"), this.GetString("email"),
	)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
	} else {
		this.sendSuccessResp(map[string]bool{"signed": v})
	}
}

// @Title List
// @Description get all the individuals by community manager
// @Param	:link_id	path 	string		true		"link id"
// @Success 200 {object} dbmodels.IndividualSigningBasicInfo
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 405 unknown_link:               unkown link id
// @Failure 406 not_yours_org:              the link doesn't belong to your community
// @Failure 500 system_error:               system error
// @router /:link_id [get]
func (this *IndividualSigningController) List() {
	action := "list individuals"
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

	r, merr := models.ListIndividualSigning(linkID, "")
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendSuccessResp(r)
}
