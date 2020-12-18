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
	if getRouterPattern(&this.Controller) == "/v1/corporation-signing/:link_id/:cla_lang/:cla_hash" {
		// signing
		this.apiPrepare("")
	} else {
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
	doWhat := "sign as corporation"
	linkID := this.GetString(":link_id")
	claLang := this.GetString(":cla_lang")

	var info models.CorporationSigningCreateOption
	if err := this.fetchInputPayload(&info); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, doWhat)
		return
	}
	info.CLALanguage = claLang
	if merr := (&info).Validate(linkID); merr != nil {
		this.sendModelErrorAsResp(merr, doWhat)
		return
	}

	orgInfo, merr := models.GetOrgOfLink(linkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, doWhat)
		return
	}

	claInfo, merr := models.GetCLAInfoSigned(linkID, claLang, dbmodels.ApplyToCorporation)
	if merr != nil {
		this.sendModelErrorAsResp(merr, doWhat)
		return
	}
	if claInfo == nil {
		// no contributor signed for this language. lock to avoid the cla to be changed
		// before writing to the db.

		unlock, err := util.Lock(genOrgFileLockPath(orgInfo.Platform, orgInfo.OrgID, orgInfo.RepoID))
		if err != nil {
			this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
			return
		}
		defer unlock()

		claInfo, merr = models.GetCLAInfoToSign(linkID, claLang, dbmodels.ApplyToCorporation)
		if merr != nil {
			this.sendModelErrorAsResp(merr, doWhat)
			return
		}
	}

	if claInfo.CLAHash != this.GetString(":cla_hash") {
		this.sendFailedResponse(400, errUnmatchedCLA, fmt.Errorf("unmatched cla"), doWhat)
		return
	}

	claFile := genCLAFilePath(linkID, dbmodels.ApplyToCorporation, claLang)
	orgSignatureFile := genOrgSignatureFilePath(linkID, claLang)
	if fr := this.checkCLAForSigning(claFile, orgSignatureFile, claInfo); fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}

	info.Info = getSingingInfo(info.Info, claInfo.Fields)

	if merr := (&info).Create(linkID); merr != nil {
		if merr.IsErrorOf(models.ErrNoLinkOrResign) {
			this.sendFailedResponse(400, errHasSigned, merr, doWhat)
		} else {
			this.sendModelErrorAsResp(merr, doWhat)
		}
		return
	}

	this.sendResponse("sign successfully", 0)

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(
		linkID, orgSignatureFile, claFile, *orgInfo, info.CorporationSigning, claInfo.Fields,
	)
}

func (this *CorporationSigningController) checkCLAForSigning(claFile, orgSignatureFile string, claInfo *dbmodels.CLAInfo) *failedResult {
	md5, err := util.Md5sumOfFile(claFile)
	if err != nil {
		return newFailedResult(500, util.ErrSystemError, err)
	}
	if md5 != claInfo.CLAHash {
		return newFailedResult(500, util.ErrSystemError, fmt.Errorf("local cla is unmatched"))
	}

	md5, err = util.Md5sumOfFile(orgSignatureFile)
	if err != nil {
		return newFailedResult(500, util.ErrSystemError, err)
	}
	if md5 != claInfo.OrgSignatureHash {
		return newFailedResult(500, util.ErrSystemError, fmt.Errorf("local org signature is unmatched"))
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
	doWhat := "resend corp signing email"
	linkID := this.GetString(":link_id")
	corpEmail := this.GetString(":email")

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}

	b, err := models.IsCorpSigningPDFUploaded(linkID, corpEmail)
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}
	if b {
		return
	}

	// TODO return cla info in GetCorpSigningDetail
	signingInfo, err := models.GetCorpSigningDetail(linkID, corpEmail)
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
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
		// TODO
		[]dbmodels.Field{},
	)

	this.sendResponse("resend email successfully", 0)
}

// @Title GetAll
// @Description get all the corporations which have signed to a org
// @router /:link_id [get]
func (this *CorporationSigningController) GetAll() {
	doWhat := "list corporation"
	linkID := this.GetString(":link_id")

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}

	r, err := models.ListCorpSignings(linkID, this.GetString("cla_language"))
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	corpMap := map[string]bool{}
	corps, err := models.ListCorpsWithPDFUploaded(linkID)
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}
	for i := range corps {
		corpMap[corps[i]] = true
	}

	type sInfo struct {
		*dbmodels.CorporationSigningSummary
		PDFUploaded bool `json:"pdf_uploaded"`
	}

	result := make([]sInfo, 0, len(r))
	for i := range r {
		items := r[i]
		result = append(result, sInfo{
			CorporationSigningSummary: &items,
			PDFUploaded:               corpMap[util.EmailSuffix(items.AdminEmail)]},
		)
	}
	this.sendResponse(result, 0)
}
