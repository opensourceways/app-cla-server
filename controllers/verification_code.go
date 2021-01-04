package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
)

type VerificationCodeController struct {
	baseController
}

func (this *VerificationCodeController) Prepare() {
	this.apiPrepare("")
}

// @Title Post
// @Description send verification code when signing
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 201 {int} map
// @Failure util.ErrSendingEmail
// @router /:org_cla_id/:email [post]
func (this *VerificationCodeController) Post() {
	sendResp := this.newFuncForSendingFailedResp("send verification code")

	orgCLAID := this.GetString(":org_cla_id")
	emailOfSigner := this.GetString(":email")

	orgCLA := &models.OrgCLA{ID: orgCLAID}
	if err := orgCLA.Get(); err != nil {
		sendResp(convertDBError1(err))
		return
	}

	code, err := models.CreateVerificationCode(
		emailOfSigner, orgCLAID, conf.AppConfig.VerificationCodeExpiry,
	)
	if err != nil {
		sendResp(parseModelError(err))
		return
	}

	this.sendSuccessResp("create verification code successfully")

	sendEmailToIndividual(
		emailOfSigner, orgCLA.OrgEmail,
		fmt.Sprintf(
			"Verification code for signing CLA on project of \"%s\"",
			orgCLA.OrgAlias,
		),
		email.VerificationCode{
			Email:      emailOfSigner,
			Org:        orgCLA.OrgAlias,
			Code:       code,
			ProjectURL: projectURL(orgCLA),
		},
	)
}
