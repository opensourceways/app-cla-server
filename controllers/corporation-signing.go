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
	if ec, err := (&info).Validate(orgCLAID); err != nil {
		this.sendFailedResponse(400, ec, err, action)
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

	err := (&info).Create(orgCLAID, orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID)
	if err != nil {
		sendResp(convertDBError1(err))
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

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		sendResp(fr)
		return
	}

	if !pl.hasOrg(org) {
		this.sendFailedResponse(400, util.ErrInvalidParameter, fmt.Errorf("can't access org:%s", org), action)
		return
	}

	opt := models.CorporationSigningListOption{
		Platform:    pl.Platform,
		OrgID:       org,
		RepoID:      this.GetString("repo_id"),
		CLALanguage: this.GetString("cla_language"),
	}

	r, err := opt.List()
	if err != nil {
		sendResp(convertDBError1(err))
		return
	}

	corpMap := map[string]bool{}
	for k := range r {
		corps, err := models.ListCorpsWithPDFUploaded(k)
		if err != nil {
			sendResp(convertDBError1(err))
			return
		}
		for i := range corps {
			corpMap[corps[i]] = true
		}
	}

	type sInfo struct {
		*dbmodels.CorporationSigningDetail
		PDFUploaded bool `json:"pdf_uploaded"`
	}

	result := map[string][]sInfo{}
	for k := range r {
		items := r[k]
		details := make([]sInfo, 0, len(items))
		for i := range items {
			details = append(details, sInfo{
				CorporationSigningDetail: &items[i],
				PDFUploaded:              corpMap[util.EmailSuffix(items[i].AdminEmail)]},
			)
		}
		result[k] = details
	}
	this.sendSuccessResp(result)
}
