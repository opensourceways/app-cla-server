package controllers

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
)

const (
	codeOfSigning = "signing"
	codeOfFindPw  = "find_pw"
)

type VerificationCodeController struct {
	baseController
}

func (this *VerificationCodeController) Prepare() {
	if strings.HasSuffix(this.routerPattern(), "/:link_id/:email/:type") {
		this.apiPrepare("")
	} else {
		this.apiPrepare(PermissionCorpAdmin)
	}
}

// @Title Post
// @Description send verification code when signing
// @Param	:link_id	path 	string					true		"link id"
// @Param	:email		path 	string					true		"email of corp"
// @param	:type		path	string					true		"type value of get code: signing or find_pw"
// @Success 201 {int} map
// @router /:link_id/:email/:type [post]
func (this *VerificationCodeController) Post() {
	action := "create verification code"
	linkID := this.GetString(":link_id")
	emailAddr := this.GetString(":email")
	typeVC := this.GetString(":type")

	orgInfo, merr := models.GetOrgOfLink(linkID)
	if merr != nil {
		this.sendFailedResponse(0, "", merr, action)
		return
	}

	code, merr := this.createCode(emailAddr, linkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	if err := sendEmailByType(typeVC, code, emailAddr, orgInfo); err != nil {
		this.sendFailedResponse(400, errUnmatchedVerificationCodeType, err, action)
	}

	this.sendSuccessResp("create verification code successfully")
}

// @Title Post
// @Description send verification code when adding email domain
// @Param	:email		path 	string		true		"email of corp"
// @Success 201 {int} map
// @Failure 400 missing_token:      token is missing
// @Failure 401 unknown_token:      token is unknown
// @Failure 402 expired_token:      token is expired
// @Failure 403 unauthorized_token: the permission of token is unauthorized
// @Failure 500 system_error:       system error
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

	code, err := this.createCode(
		corpEmail, models.PurposeOfAddingEmailDomain(pl.Email),
	)
	if err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("create verification code successfully")

	sendEmailToIndividual(
		corpEmail, pl.OrgEmail,
		"Verification code for adding corporation's another email domain",
		email.AddingCorpEmailDomain{
			Corp:       pl.Corp,
			Org:        pl.OrgAlias,
			Code:       code,
			ProjectURL: pl.OrgInfo.ProjectURL(),
		},
	)
}

func (this *VerificationCodeController) createCode(to, purpose string) (string, models.IModelError) {
	return models.CreateVerificationCode(
		to, purpose, config.AppConfig.VerificationCodeExpiry,
	)
}

func sendEmailByType(typeVC, code, emailAddr string, orgInfo *models.OrgInfo) error {
	var err error
	switch typeVC {
	case codeOfSigning:
		sendEmailToIndividual(
			emailAddr, orgInfo.OrgEmail,
			fmt.Sprintf(
				"Verification code for signing CLA on project of \"%s\"",
				orgInfo.OrgAlias,
			),
			email.VerificationCode{
				Email:      emailAddr,
				Org:        orgInfo.OrgAlias,
				Code:       code,
				ProjectURL: orgInfo.ProjectURL(),
			},
		)
	case codeOfFindPw:
		sendEmailToIndividual(
			emailAddr, orgInfo.OrgEmail,
			"Verification code for retrieve password ",
			email.FindPasswordVerifyCode{
				Email:      emailAddr,
				Org:        orgInfo.OrgAlias,
				Code:       code,
				ProjectURL: orgInfo.ProjectURL(),
			},
		)
	default:
		err = fmt.Errorf(errUnmatchedVerificationCodeType)
	}
	return err
}
