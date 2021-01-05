package controllers

import (
	"fmt"

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
	}
}

// @Title Post
// @Description sign as individual
// @Param	:org_cla_id	path 	string				true		"org cla id"
// @Param	body		body 	models.IndividualSigning	true		"body for individual signing"
// @Success 201 {int} map
// @Failure util.ErrHasSigned
// @router /:org_cla_id [post]
func (this *IndividualSigningController) Post() {
	action := "sign individual cla"
	sendResp := this.newFuncForSendingFailedResp(action)

	orgCLAID := this.GetString(":org_cla_id")
	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		sendResp(fr)
		return
	}

	var info models.IndividualSigning
	if fr := this.fetchInputPayload(&info); fr != nil {
		sendResp(fr)
		return
	}
	if err := (&info).Validate(pl.Email); err != nil {
		sendResp(parseModelError(err))
		return
	}

	orgCLA := &models.OrgCLA{ID: orgCLAID}
	if err := orgCLA.Get(); err != nil {
		sendResp(convertDBError1(err))
		return
	}
	if isNotIndividualCLA(orgCLA) {
		this.sendFailedResponse(400, util.ErrInvalidParameter, fmt.Errorf("invalid cla"), action)
		return
	}

	cla := &models.CLA{ID: orgCLA.CLAID}
	if err := cla.GetFields(); err != nil {
		sendResp(convertDBError1(err))
		return
	}

	info.Info = getSingingInfo(info.Info, cla.Fields)

	if err := (&info).Create(orgCLAID, true); err != nil {
		if err.IsErrorOf(models.ErrNoLinkOrResigned) {
			this.sendFailedResponse(400, errResigned, err, action)
		} else {
			sendResp(parseModelError(err))
		}
		return
	}

	this.sendSuccessResp("sign successfully")
}

// @Title Check
// @Description check whether contributor has signed cla
// @Param	platform	path 	string	true		"code platform"
// @Param	org		path 	string	true		"org"
// @Param	repo		path 	string	true		"repo"
// @Param	email		query 	string	true		"email"
// @Success 200
// @router /:platform/:org_repo [get]
func (this *IndividualSigningController) Check() {
	sendResp := this.newFuncForSendingFailedResp("check individual signing")
	org, repo := parseOrgAndRepo(this.GetString(":org_repo"))
	emailOfSigner := this.GetString("email")

	opt := models.OrgCLAListOption{
		Platform: this.GetString(":platform"),
		OrgID:    org,
		RepoID:   repo,
		ApplyTo:  dbmodels.ApplyToIndividual,
	}
	signings, err := opt.List()
	if err != nil {
		sendResp(convertDBError1(err))
		return
	}
	if len(signings) == 0 {
		return
	}

	if v, merr := models.IsIndividualSigned(signings[0].ID, emailOfSigner); merr != nil {
		sendResp(parseModelError(merr))
	} else {
		this.sendSuccessResp(map[string]bool{"signed": v})
	}
}
