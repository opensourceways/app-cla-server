package controllers

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
)

type VerificationCodeController struct {
	baseController
}

func (this *VerificationCodeController) Prepare() {
	if strings.HasSuffix(this.routerPattern(), "/:link_id/:email") {
		this.apiPrepare("")
	} else {
		// get, update and delete employee
		this.apiPrepare(PermissionCorpAdmin)
	}
}

// @Title Post
// @Description send verification code when signing
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 201 {int} map
// @Failure util.ErrSendingEmail
// @router /:link_id/:email [post]
func (this *VerificationCodeController) Post() {
	action := "create verification code"
	linkID := this.GetString(":link_id")
	emailOfSigner := this.GetString(":email")

	orgInfo, merr := models.GetOrgOfLink(linkID)
	if merr != nil {
		this.sendFailedResponse(0, "", merr, action)
		return
	}

	code, err := models.CreateVerificationCode(
		emailOfSigner, linkID, config.AppConfig.VerificationCodeExpiry,
	)
	if err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("create verification code successfully")

	sendEmailToIndividual(
		emailOfSigner, orgInfo.OrgEmail,
		fmt.Sprintf(
			"Verification code for signing CLA on project of \"%s\"",
			orgInfo.OrgAlias,
		),
		email.VerificationCode{
			Email:      emailOfSigner,
			Org:        orgInfo.OrgAlias,
			Code:       code,
			ProjectURL: orgInfo.ProjectURL(),
		},
	)
}

// @Title Post
// @Description send verification code when adding email domain
// @Param	:email		path 	string		true		"email of corp"
// @Success 201 {int} map
// @Failure util.ErrSendingEmail
// @router /:email [post]
func (this *VerificationCodeController) EmailDomain() {
	action := "create verification code for adding email domain"
	sendResp := this.newFuncForSendingFailedResp(action)
	corpEmail := this.GetString(":email")

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	code, err := models.CreateVerificationCode(
		corpEmail, pl.Email, config.AppConfig.VerificationCodeExpiry,
	)
	if err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("create verification code successfully")

	sendEmailToIndividual(
		corpEmail, pl.OrgEmail,
		"Verification code for adding another email domain",
		email.AddingCorpEmailDomain{
			Corp:       pl.Corp,
			Org:        pl.OrgAlias,
			Code:       code,
			ProjectURL: pl.OrgInfo.ProjectURL(),
		},
	)
}
