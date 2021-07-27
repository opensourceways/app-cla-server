package controllers

import (
	"fmt"
	"net/url"
	"path"
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
		this.apiPrepare(PermissionCorpAdmin)
	}
}

// @Title Post
// @Description send verification code when signing
// @Param	:link_id	path 	string					true		"link id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 201 {int} map
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

	code, err := this.createCode(emailOfSigner, linkID)
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

//@Title CodeWithFindPwd
//@Description send verification code when find password
//@Param link_id	path 	string	true	"link id"
//@Param email		path	string	true	"email of contributor"
//@Success 201 {int} map
//@Failure 400 util.ErrSendingEmail
//@router /password_retrieve/:link_id/:email [post]
func (this *VerificationCodeController) PasswordRetrieve() {
	action := "send email with find password"
	linkID := this.GetString(":link_id")

	orgInfo, err := models.GetOrgOfLink(linkID)
	if err != nil {
		this.sendFailedResponse(400, string(models.ErrNoLinkOrNoManager), err, action)
		return
	}

	cmEmail := this.GetString(":email")
	code, err := models.CreateVerificationCode(
		cmEmail, linkID, config.AppConfig.VerificationCodeExpiry,
	)
	if err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	rpw := models.RetrievePW{LinkID: linkID, Email: cmEmail, Code: code}
	ens, mErr := rpw.Encrypt()
	if mErr != nil {
		this.sendModelErrorAsResp(mErr, action)
	}
	URL, err1 := genPasswordRetrieveURL(linkID, ens)
	if err1 != nil {
		this.sendFailedResponse(400, errSystemError, err1, action)
	}
	this.sendSuccessResp("send the retrieve password email success")

	sendEmailToIndividual(
		cmEmail, orgInfo.OrgEmail,
		"[CLA] Please reset your password",
		email.FindPasswordVerifyCode{URL: URL},
	)
}

func genPasswordRetrieveURL(linkID, encryptParam string) (string, error) {
	URL, err := url.Parse(config.AppConfig.CLAPlatformURL)
	if err != nil {
		return "", nil
	}

	URL.Path = path.Join("password-retrieve", linkID, encryptParam)
	return URL.String(), nil
}
