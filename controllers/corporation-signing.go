package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
	"github.com/opensourceways/app-cla-server/worker"
)

type CorporationSigningController struct {
	baseController
}

func (this *CorporationSigningController) Prepare() {
	if this.routerPattern() == "/v1/corporation-signing/:link_id/:cla_lang/:cla_hash" {
		this.apiPrepare("")
	} else {
		// not signing
		this.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Post
// @Description sign as corporation
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Param	body		body 	models.CorporationSigningCreateOption	true		"body for corporation signing"
// @Success 201 {int} map
// @Failure util.ErrHasSigned
// @Failure util.ErrWrongVerificationCode
// @Failure util.ErrVerificationCodeExpired
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

			claFile := genCLAFilePath(linkID, dbmodels.ApplyToCorporation, claLang)
			orgSignatureFile := genOrgSignatureFilePath(linkID, claLang)
			if fr := this.checkCLAForSigning(claFile, orgSignatureFile, claInfo); fr != nil {
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
				linkID, orgSignatureFile, claFile, *orgInfo,
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

func (this *CorporationSigningController) checkCLAForSigning(claFile, orgSignatureFile string, claInfo *dbmodels.CLAInfo) *failedApiResult {
	md5, err := util.Md5sumOfFile(claFile)
	if err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}
	if md5 != claInfo.CLAHash {
		return newFailedApiResult(500, errSystemError, fmt.Errorf("local cla is unmatched"))
	}

	md5, err = util.Md5sumOfFile(orgSignatureFile)
	if err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}
	if md5 != claInfo.OrgSignatureHash {
		return newFailedApiResult(500, errSystemError, fmt.Errorf("local org signature is unmatched"))
	}
	return nil
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

	fields, signingInfo, merr := models.GetCorpSigningDetail(linkID, corpEmail)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}
	if fields == nil {
		this.sendFailedResponse(400, errUnsigned, fmt.Errorf("no data"), action)
		return
	}

	claFile := genCLAFilePath(linkID, dbmodels.ApplyToCorporation, signingInfo.CLALanguage)
	orgSignatureFile := genOrgSignatureFilePath(linkID, signingInfo.CLALanguage)

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(
		linkID, orgSignatureFile, claFile, *pl.orgInfo(linkID),
		models.CorporationSigning{
			CorporationSigningBasicInfo: signingInfo.CorporationSigningBasicInfo,
			Info:                        signingInfo.Info,
		},
		fields,
	)

	this.sendSuccessResp("resend email successfully")
}

// @Title GetAll
// @Description get all the corporations which have signed to a org
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

	type sInfo struct {
		*dbmodels.CorporationSigningSummary
		PDFUploaded bool `json:"pdf_uploaded"`
	}

	details := make([]sInfo, 0, len(r))
	for k := range r {
		details = append(details, sInfo{
			CorporationSigningSummary: &r[k],
			PDFUploaded:               pdfMap[util.EmailSuffix(r[k].AdminEmail)]},
		)
	}
	this.sendSuccessResp(details)
}
