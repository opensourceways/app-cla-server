/*
 * Copyright (C) 2021. Huawei Technologies Co., Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"fmt"
	"strings"

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

	b, merr := models.IsCorpSigned(linkID, info.Email)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}
	if b {
		this.sendFailedResponse(400, errGoToSignEmployeeCLA, fmt.Errorf("sign employee cla instead"), action)
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
// @Failure 400 no_link:      there is not link for this org and repo
// @Failure 500 system_error: system error
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

	r, merr := models.ListIndividualSigning(linkID, "", "")
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	corps, merr := models.ListCorpSignings(linkID, "")
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	m := make(map[string]bool, len(corps))
	for i := range corps {
		if corps[i].AdminAdded {
			m[util.EmailSuffix(corps[i].AdminEmail)] = true
		}
	}

	result := make([]*dbmodels.IndividualSigningBasicInfo, 0, len(r))
	for i := range r {
		if !m[util.EmailSuffix(r[i].Email)] {
			result = append(result, &r[i])
		}
	}
	this.sendSuccessResp(result)
}
