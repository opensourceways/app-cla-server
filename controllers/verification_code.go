package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
)

type VerificationCodeController struct {
	baseController
}

func (this *VerificationCodeController) Prepare() {
	this.stopRunIfSignSerivceIsUnabled()

	if strings.HasSuffix(this.routerPattern(), "/:link_id/:email") {
		this.apiPrepare("")
	} else {
		this.apiPrepare(PermissionCorpAdmin)
	}
}

// @Title Post
// @Description send verification code when signing
// @Param  link_id  path  string                               true  "link id"
// @Param  body     body  controllers.verificationCodeRequest  true  "body for verification code"
// @Success 201 {int} map
// @router /:link_id [post]
func (this *VerificationCodeController) Post() {
	action := "create verification code"
	linkID := this.GetString(":link_id")

	var req verificationCodeRequest
	if fr := this.fetchInputPayload(&req); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if err := req.validate(); err != nil {
		this.sendFailedResultAsResp(
			newFailedApiResult(400, errParsingApiBody, err),
			action,
		)
		return
	}

	if !emailLimiter.check(linkID, req.Email) {
		this.sendFailedResponse(
			http.StatusBadRequest, errTooManyRequest,
			fmt.Errorf("too many request"), action,
		)

		return
	}

	orgInfo, merr := models.GetOrgOfLink(linkID)
	if merr != nil {
		this.sendFailedResponse(0, "", merr, action)
		return
	}

	code, err := models.CreateCodeForSigning(linkID, req.Email)
	if err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("create verification code successfully")

	sendEmailToIndividual(
		req.Email, orgInfo.OrgEmail,
		fmt.Sprintf(
			"Verification code for signing CLA on project of \"%s\"",
			orgInfo.OrgAlias,
		),
		email.VerificationCode{
			Email:      req.Email,
			Org:        orgInfo.OrgAlias,
			Code:       code,
			ProjectURL: orgInfo.ProjectURL(),
		},
	)
}

// @Title Post
// @Description send verification code when adding email domain
// @Param  body  body  controllers.verificationCodeRequest  true  "body for verification code"
// @Success 201 {int} map
// @Failure 400 missing_token:      token is missing
// @Failure 401 unknown_token:      token is unknown
// @Failure 402 expired_token:      token is expired
// @Failure 403 unauthorized_token: the permission of token is unauthorized
// @Failure 500 system_error:       system error
// @router / [post]
func (this *VerificationCodeController) EmailDomain() {
	action := "create verification code for adding email domain"
	sendResp := this.newFuncForSendingFailedResp(action)

	var req verificationCodeRequest
	if fr := this.fetchInputPayload(&req); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if err := req.validate(); err != nil {
		this.sendFailedResultAsResp(
			newFailedApiResult(400, errParsingApiBody, err),
			action,
		)
		return
	}

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	orgInfo, merr := models.GetOrgOfLink(pl.LinkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)

		return
	}

	// TODO csid
	code, err := models.CreateCodeForAddingEmailDomain("", req.Email)
	if err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("create verification code successfully")

	sendEmailToIndividual(
		req.Email, orgInfo.OrgEmail,
		"Verification code for adding corporation's another email domain",
		email.AddingCorpEmailDomain{
			Corp:       pl.Corp,
			Org:        orgInfo.OrgAlias,
			Code:       code,
			ProjectURL: orgInfo.ProjectURL(),
		},
	)
}

type verificationCodeRequest struct {
	Email string `json:"email" required:"true"`
}

func (req *verificationCodeRequest) validate() error {
	if req.Email == "" {
		return errors.New("missing email")
	}

	return nil
}
