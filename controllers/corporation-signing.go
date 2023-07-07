package controllers

import (
	"strings"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/worker"
)

type CorporationSigningController struct {
	baseController
}

func (this *CorporationSigningController) Prepare() {
	v := this.routerPattern()
	if strings.HasSuffix(v, ":cla_hash") || strings.HasSuffix(v, ":link_id/corps/:email") {
		this.apiPrepare("")
	} else {
		// not signing
		this.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Post
// @Description sign corporation cla
// @Param	:link_id	path 	string					true		"link id"
// @Param	:cla_lang	path 	string					true		"cla language"
// @Param	:cla_hash	path 	string					true		"the hash of cla content"
// @Param	body		body 	models.CorporationSigningCreateOption	true		"body for signing corporation cla"
// @Success 201 {string} "sign successfully"
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 error_parsing_api_body:     parse input paraemter failed
// @Failure 402 expired_verification_code:  the verification code is expired
// @Failure 403 wrong_verification_code:    the verification code is wrong
// @Failure 404 not_an_email:               the email inputed is wrong
// @Failure 405 no_link:                    the link id is not exists
// @Failure 406 unmatched_cla:              the cla hash is not equal to the one of backend server
// @Failure 407 resigned:                   the signer has signed the cla
// @Failure 500 system_error:               system error
// @router /:link_id/:cla_lang/:cla_id [post]
func (this *CorporationSigningController) Post() {
	action := "sign as corporation"
	linkID := this.GetString(":link_id")
	claLang := this.GetString(":cla_lang")
	claId := this.GetString(":cla_id")

	var info models.CorporationSigningCreateOption
	if fr := this.fetchInputPayload(&info); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	info.CLALanguage = claLang

	if err := info.Validate(linkID); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	orgInfo, claInfo, merr := models.GetLinkCLA(linkID, claId)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	info.Info = getSingingInfo(info.Info, claInfo.Fields)
	info.CLAId = claId

	if err := models.SignCropCLA(linkID, &info); err != nil {
		if err.IsErrorOf(models.ErrNoLinkOrResigned) {
			this.sendFailedResponse(400, errResigned, err, action)
		} else {
			this.sendModelErrorAsResp(err, action)
		}

		return
	}

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(
		linkID, &orgInfo, &claInfo, &info.CorporationSigning,
	)

	this.sendSuccessResp("sign successfully")
}

// @Title Delete
// @Description delete corp signing
// @Param  link_id     path  string  true  "link id"
// @Param  signing_id  path  string  true  "corp signing id"
// @Success 204 {string} delete success!
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 405 not_yours_org:              the link doesn't belong to your community
// @Failure 406 unknown_link:               unkown link id
// @Failure 407 no_link:                    the link id is not exists
// @Failure 500 system_error:               system error
// @router /:link_id/:signing_id [delete]
func (this *CorporationSigningController) Delete() {
	action := "delete corp signing"
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

	csId := this.GetString(":signing_id")
	if err := models.RemoveCorpSigning(csId); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("delete corp signing successfully")
}

// @Title ResendCorpSigningEmail
// @Description resend corp signing email
// @Param  link_id      path  string  true  "link id"
// @Param  signing_id  path  string  true  "corp email"
// @Success 201 {int} map
// @router /:link_id/:signing_id [post]
func (this *CorporationSigningController) ResendCorpSigningEmail() {
	action := "resend corp signing email"
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

	signingInfo, merr := models.GetCorpSigning(this.GetString(":signing_id"))
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	orgInfo, claInfo, merr := models.GetLinkCLA(linkID, signingInfo.CLAId)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(
		linkID, &orgInfo, &claInfo, &signingInfo,
	)

	this.sendSuccessResp("resend email successfully")
}

type corpsSigningResult struct {
	*dbmodels.CorporationSigningSummary
	PDFUploaded bool `json:"pdf_uploaded"`
}

// @Title GetAll
// @Description get all the corporations which have signed to a org
// @Param  link_id  path  string  true  "link id"
// @Success 200 {object} models.CorporationSigningSummary
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 405 unknown_link:               unkown link id
// @Failure 406 not_yours_org:              the link doesn't belong to your community
// @Failure 500 system_error:               system error
// @router /:link_id [get]
func (this *CorporationSigningController) GetAll() {
	action := "list corporation"
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

	if r, merr := models.ListCorpSigning(linkID); merr != nil {
		this.sendModelErrorAsResp(merr, action)
	} else {
		this.sendSuccessResp(r)
	}
}

// @Title GetAll
// @Description get all the corporations which have been deleted
// @Param	:link_id	path 	string		true		"link id"
// @Success 200 {object} dbmodels.CorporationSigningBasicInfo
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 405 unknown_link:               unkown link id
// @Failure 406 not_yours_org:              the link doesn't belong to your community
// @Failure 500 system_error:               system error
// @router /deleted/:link_id [get]
func (this *CorporationSigningController) ListDeleted() {
	this.sendSuccessResp(nil)
}

// @Title GetCorpInfo
// @Description get all the corporations by email
// @Param  link_id  path  string  true  "link id"
// @Param  email    path  string  true  "email"
// @Success 200 {object} interface{}
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 unknown_link:               unkown link id
// @Failure 500 system_error:               system error
// @router /:link_id/corps/:email [get]
func (this *CorporationSigningController) GetCorpInfo() {
	action := "list corporation info"

	r, merr := models.FindCorpSummary(
		this.GetString(":link_id"), this.GetString(":email"),
	)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
	} else {
		this.sendSuccessResp(r)
	}
}
