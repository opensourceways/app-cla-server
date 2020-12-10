package controllers

import (
	"fmt"
	"net/http"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type IndividualSigningController struct {
	baseController
}

func (this *IndividualSigningController) Prepare() {
	if getRequestMethod(&this.Controller) == http.MethodPost {
		// sign as individual
		this.apiPrepare(PermissionIndividualSigner)
	} else {
		// check sign
		this.apiPrepare("")
	}
}

// @Title Post
// @Description sign as individual
// @Param	:org_cla_id	path 	string				true		"org cla id"
// @Param	body		body 	models.IndividualSigning	true		"body for individual signing"
// @Success 201 {int} map
// @Failure util.ErrHasSigned
// @router /:link_id/:cla_lang/:cla_hash [post]
func (this *IndividualSigningController) Post() {
	doWhat := "sign as individual"

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	var info models.IndividualSigning
	if err := this.fetchInputPayload(&info); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, doWhat)
		return
	}
	if ec, err := (&info).Validate(pl.Email); err != nil {
		this.sendFailedResponse(400, ec, err, doWhat)
		return
	}

	linkID := this.GetString(":link_id")
	claLang := this.GetString(":cla_lang")
	claInfo, err := models.GetCLAInfoSigned(linkID, claLang, dbmodels.ApplyToIndividual)
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}
	if claInfo == nil {
		// TODO get cla info again under lock
	}
	if claInfo.CLAHash != this.GetString(":cla_hash") {
		this.sendFailedResponse(400, util.ErrInvalidParameter, fmt.Errorf("invalid cla"), doWhat)
		return
	}

	info.Info = getSingingInfo(info.Info, claInfo.Fields)

	err = (&info).Create(linkID, true)
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	this.sendResponse("sign successfully", 0)
}

// @Title Check
// @Description check whether contributor has signed cla
// @Param	platform	path 	string	true		"code platform"
// @Param	org		path 	string	true		"org"
// @Param	repo		path 	string	true		"repo"
// @Param	email		query 	string	true		"email"
// @Success 200
// @router /:platform/:org/:repo [get]
func (this *IndividualSigningController) Check() {
	doWhat := "check individual signing"

	v, err := models.IsIndividualSigned(
		buildOrgRepo(this.GetString(":platform"), this.GetString(":org"), this.GetString(":repo")),
		this.GetString("email"),
	)
	if err != nil {
		statusCode, errCode := convertDBError(err)
		if errCode != util.ErrHasNotSigned {
			this.sendFailedResponse(statusCode, errCode, err, doWhat)
			return
		}
		v = false
	}

	this.sendResponse(map[string]bool{"signed": v}, 0)
}
