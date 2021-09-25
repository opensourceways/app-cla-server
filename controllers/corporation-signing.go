package controllers

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
	"github.com/opensourceways/app-cla-server/worker"
)

type CorporationSigningController struct {
	baseController
}

func (this *CorporationSigningController) Prepare() {
	if isSigningServiceNotStarted() {
		this.StopRun()
	}

	if strings.HasSuffix(this.routerPattern(), ":cla_hash") {
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
// @router /:link_id/:cla_lang/:cla_hash [post]
func (this *CorporationSigningController) Post() {
	action := "sign as corporation"
	linkID := this.GetString(":link_id")
	claLang := this.GetString(":cla_lang")

	var info models.CorporationSigningCreateOption
	if fr := this.fetchInputPayload(&info); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	info.CLALanguage = claLang

	if err := (&info).Validate(linkID); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	orgInfo, merr := models.GetOrgOfLink(linkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	fr := signHelper(
		linkID, claLang, dbmodels.ApplyToCorporation,
		func(claInfo *models.CLAInfo) *failedApiResult {
			if claInfo.CLAHash != this.GetString(":cla_hash") {
				return newFailedApiResult(400, errUnmatchedCLA, fmt.Errorf("unmatched cla"))
			}

			claFile := genCLAFilePath(linkID, dbmodels.ApplyToCorporation, claLang, claInfo.CLAHash)
			if fr := this.checkCLAForSigning(claFile, claInfo); fr != nil {
				return fr
			}

			info.Info = getSingingInfo(info.Info, claInfo.Fields)

			if err := (&info).Create(linkID); err != nil {
				if err.IsErrorOf(models.ErrNoLinkOrResigned) {
					return newFailedApiResult(400, errResigned, err)
				}
				return parseModelError(err)
			}

			worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(
				linkID, claFile, *orgInfo,
				info.CorporationSigning, claInfo.Fields,
			)

			return nil
		},
	)
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
	} else {
		this.sendSuccessResp("sign successfully")
	}
}

func (this *CorporationSigningController) checkCLAForSigning(claFile string, claInfo *dbmodels.CLAInfo) *failedApiResult {
	md5, err := util.Md5sumOfFile(claFile)
	if err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}
	if md5 != claInfo.CLAHash {
		return newFailedApiResult(500, errSystemError, fmt.Errorf("local cla is unmatched"))
	}

	return nil
}

// @Title Delete
// @Description delete corp signing
// @Param	:link_id	path 	string		true		"link id"
// @Param	:email		path 	string		true		"corp email"
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
// @router /:link_id/:email [delete]
func (this *CorporationSigningController) Delete() {
	action := "delete corp signing"
	linkID := this.GetString(":link_id")
	corpEmail := this.GetString(":email")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	unlock, fr := lockOnRepo(pl.orgInfo(linkID))
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	defer unlock()

	managers, merr := models.ListCorporationManagers(linkID, corpEmail, dbmodels.RoleAdmin)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}
	if len(managers) > 0 {
		this.sendFailedResponse(
			400, errCorpManagerExists,
			fmt.Errorf("can't delete corp signing info, because admin manager exists"), action)
		return
	}

	if err := models.DeleteCorpSigning(linkID, corpEmail); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("delete corp signing successfully")
}

// @Title ResendCorpSigningEmail
// @Description resend corp signing email
// @Param	:org_id		path 	string		true		"org cla id"
// @Param	:email		path 	string		true		"corp email"
// @Success 201 {int} map
// @router /:link_id/:email [post]
func (this *CorporationSigningController) ResendCorpSigningEmail() {
	action := "resend corp signing email"
	linkID := this.GetString(":link_id")
	corpEmail := this.GetString(":email")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	claInfo, signingInfo, merr := models.GetCorpSigningDetail(linkID, corpEmail)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}
	if claInfo == nil {
		this.sendFailedResponse(400, errUnsigned, fmt.Errorf("no data"), action)
		return
	}

	claFile := genCLAFilePath(linkID, dbmodels.ApplyToCorporation, signingInfo.CLALanguage, claInfo.CLAHash)

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(
		linkID, claFile, *pl.orgInfo(linkID),
		models.CorporationSigning{
			CorporationSigningBasicInfo: signingInfo.CorporationSigningBasicInfo,
			Info:                        signingInfo.Info,
		},
		claInfo.Fields,
	)

	this.sendSuccessResp("resend email successfully")
}

type corpsSigningResult struct {
	*dbmodels.CorporationSigningSummary
	PDFUploaded bool `json:"pdf_uploaded"`
}

// @Title GetAll
// @Description get all the corporations which have signed to a org
// @Param	:link_id	path 	string		true		"link id"
// @Success 200 {object} controllers.corpsSigningResult
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

	r, merr := models.ListCorpSignings(linkID, this.GetString("cla_language"))
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}
	if len(r) == 0 {
		this.sendSuccessResp(nil)
		return
	}

	pdfs, err := models.ListCorpsWithPDFUploaded(linkID)
	if err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}
	pdfMap := map[string]bool{}
	for i := range pdfs {
		pdfMap[pdfs[i]] = true
	}

	details := make([]corpsSigningResult, 0, len(r))
	for k := range r {
		details = append(details, corpsSigningResult{
			CorporationSigningSummary: &r[k],
			PDFUploaded:               pdfMap[util.EmailSuffix(r[k].AdminEmail)]},
		)
	}
	this.sendSuccessResp(details)
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
	action := "list deleted corporations"
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

	r, merr := models.ListDeletedCorpSignings(linkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendSuccessResp(r)
}
