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
// @router /:link_id/:email [post]
func (this *VerificationCodeController) Post() {
	action := "send verification code"
	linkID := this.GetString(":link_id")
	inputEmail := this.GetString(":email")

	orgInfo, err := models.GetOrgOfLink(linkID)
	if err != nil {
		this.sendFailedResponse(0, "", err, action)
		return
	}

	code, err := models.CreateVerificationCode(
		inputEmail, linkID, conf.AppConfig.VerificationCodeExpiry,
	)
	if err != nil {
		this.sendFailedResponse(0, "", err, action)
		return
	}

	this.sendResponse("create verification code successfully", 0)

	sendEmailToIndividual(
		inputEmail, orgInfo.OrgEmail,
		fmt.Sprintf(
			"Verification code for signing CLA on project of \"%s\"",
			orgInfo.OrgAlias,
		),
		email.VerificationCode{
			Email:      inputEmail,
			Org:        orgInfo.OrgAlias,
			Code:       code,
			ProjectURL: orgInfo.ProjectURL(),
		},
	)
}
