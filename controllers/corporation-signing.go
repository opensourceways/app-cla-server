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
	if this.routerPattern() == "/v1/corporation-signing/:org_cla_id" {
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
// @router /:org_cla_id [post]
func (this *CorporationSigningController) Post() {
	action := "sign as corporation"
	sendResp := this.newFuncForSendingFailedResp(action)
	orgCLAID := this.GetString(":org_cla_id")

	var info models.CorporationSigningCreateOption
	if fr := this.fetchInputPayload(&info); fr != nil {
		sendResp(fr)
		return
	}
	if err := (&info).Validate(orgCLAID); err != nil {
		sendResp(parseModelError(err))
		return
	}

	orgCLA := &models.OrgCLA{ID: orgCLAID}
	if err := orgCLA.Get(); err != nil {
		sendResp(convertDBError1(err))
		return
	}
	if isNotCorpCLA(orgCLA) {
		this.sendFailedResponse(400, util.ErrInvalidParameter, fmt.Errorf("invalid cla"), action)
		return
	}

	cla := &models.CLA{ID: orgCLA.CLAID}
	if err := cla.Get(); err != nil {
		sendResp(convertDBError1(err))
		return
	}

	info.Info = getSingingInfo(info.Info, cla.Fields)

	if err := (&info).Create(orgCLAID); err != nil {
		if err.IsErrorOf(models.ErrNoLinkOrResigned) {
			this.sendFailedResponse(400, errResigned, err, action)
		} else {
			sendResp(parseModelError(err))
		}
		return
	}

	this.sendSuccessResp("sign successfully")

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(orgCLA, &info.CorporationSigning, cla)
}

// @Title ResendCorpSigningEmail
// @Description resend corp signing email
// @Param	:org_id		path 	string		true		"org cla id"
// @Param	:email		path 	string		true		"corp email"
// @Success 201 {int} map
// @router /:org_id/:email [post]
func (this *CorporationSigningController) ResendCorpSigningEmail() {
	action := "resend corp signing email"
	sendResp := this.newFuncForSendingFailedResp(action)

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		sendResp(fr)
		return
	}

	org, repo := parseOrgAndRepo(this.GetString(":org_id"))
	if !pl.hasOrg(org) {
		this.sendFailedResponse(400, util.ErrInvalidParameter, fmt.Errorf("can't access org:%s", org), action)
		return
	}

	orgCLAID, signingInfo, err := models.GetCorpSigningInfo(
		pl.Platform, org, repo, this.GetString(":email"),
	)
	if err != nil {
		sendResp(convertDBError1(err))
		return
	}

	orgCLA := &models.OrgCLA{ID: orgCLAID}
	if err := orgCLA.Get(); err != nil {
		sendResp(convertDBError1(err))
		return
	}

	cla := &models.CLA{ID: orgCLA.CLAID}
	if err := cla.Get(); err != nil {
		sendResp(convertDBError1(err))
		return
	}

	this.sendSuccessResp("resend email successfully")

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(
		orgCLA, (*models.CorporationSigning)(signingInfo), cla,
	)

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
	if r == nil {
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
