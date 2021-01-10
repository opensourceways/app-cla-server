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

	claInfo, fr := getCLAInfoSigned(linkID, claLang, dbmodels.ApplyToCorporation)
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if claInfo == nil {
		// no contributor signed for this language. lock to avoid the cla to be changed
		// before writing to the db.

		unlock, err := util.Lock(genOrgFileLockPath(orgInfo.Platform, orgInfo.OrgID, orgInfo.RepoID))
		if err != nil {
			this.sendFailedResponse(500, util.ErrSystemError, err, action)
			return
		}
		defer unlock()

		claInfo, merr = models.GetCLAInfoToSign(linkID, claLang, dbmodels.ApplyToCorporation)
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
		this.sendFailedResponse(400, errUnmatchedCLA, fmt.Errorf("unmatched cla"), action)
		return
	}

	claFile := genCLAFilePath(linkID, dbmodels.ApplyToCorporation, claLang)
	orgSignatureFile := genOrgSignatureFilePath(linkID, claLang)
	if fr := this.checkCLAForSigning(claFile, orgSignatureFile, claInfo); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	info.Info = getSingingInfo(info.Info, claInfo.Fields)

	if err := (&info).Create(linkID); err != nil {
		if err.IsErrorOf(models.ErrNoLinkOrResigned) {
			this.sendFailedResponse(400, errResigned, err, action)
		} else {
			this.sendModelErrorAsResp(err, action)
		}
		return
	}

	this.sendSuccessResp("sign successfully")

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(
		linkID, orgSignatureFile, claFile, *orgInfo, info.CorporationSigning, claInfo.Fields)
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
// @router /:org_id [get]
func (this *CorporationSigningController) GetAll() {
	action := "list corporation"
	sendResp := this.newFuncForSendingFailedResp(action)
	org := this.GetString(":org_id")
	repo := this.GetString("repo_id")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		sendResp(fr)
		return
	}
	if !pl.hasOrg(org) {
		this.sendFailedResponse(400, util.ErrInvalidParameter, fmt.Errorf("can't access org:%s", org), action)
		return
	}

	linkID, fr := getLinkID(pl.Platform, org, repo, dbmodels.ApplyToCorporation)
	if fr != nil {
		sendResp(fr)
		return
	}

	r, merr := models.ListCorpSignings(linkID, this.GetString("cla_language"))
	if merr != nil {
		sendResp(parseModelError(merr))
		return
	}
	if len(r) == 0 {
		this.sendSuccessResp(map[string]bool{})
		return
	}

	pdfs, err := models.ListCorpsWithPDFUploaded(linkID)
	if err != nil {
		sendResp(convertDBError1(err))
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
		items := r[k]
		details = append(details, sInfo{
			CorporationSigningSummary: &items,
			PDFUploaded:               pdfMap[util.EmailSuffix(items.AdminEmail)]},
		)
	}
	this.sendSuccessResp(map[string][]sInfo{linkID: details})
}
