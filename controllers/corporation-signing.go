package controllers

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
	"github.com/opensourceways/app-cla-server/worker"
)

type CorporationSigningController struct {
	baseController
}

func (this *CorporationSigningController) Prepare() {
	if getRouterPattern(&this.Controller) == "/v1/corporation-signing/:link_id" {
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
// @router /:link_id [post]
func (this *CorporationSigningController) Post() {
	doWhat := "sign as corporation"
	linkID := this.GetString(":link_id")

	var info models.CorporationSigningCreateOption
	if err := this.fetchInputPayload(&info); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, doWhat)
		return
	}
	if ec, err := (&info).Validate(linkID); err != nil {
		this.sendFailedResponse(400, ec, err, doWhat)
		return
	}

	orgCLA := &models.OrgCLA{ID: linkID}
	if err := orgCLA.Get(); err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	cla := &models.CLA{ID: orgCLA.CLAID}
	if err := cla.Get(); err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	info.Info = getSingingInfo(info.Info, cla.Fields)

	if err := (&info).Create(linkID); err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	this.sendResponse("sign successfully", 0)

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(orgCLA, &info.CorporationSigning, cla)
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

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}
	if !pl.hasLink(linkID) {
		// TODO
	}

	signingInfo, err := models.GetCorpSigningDetail(linkID, this.GetString(":email"))
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	orgCLA := &models.OrgCLA{ID: linkID}
	if err := orgCLA.Get(); err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	cla := &models.CLA{ID: orgCLA.CLAID}
	if err := cla.Get(); err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(
		orgCLA, (*models.CorporationSigning)(signingInfo), cla,
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
	if !pl.hasLink(linkID) {
		// TODO
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
