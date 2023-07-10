package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
)

type VerificationCodeController struct {
	baseController
}

func (this *VerificationCodeController) Prepare() {
	this.apiPrepare("")
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

	orgInfo, merr := models.GetLink(linkID)
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
		req.Email, &orgInfo,
		fmt.Sprintf(
			"Verification code for signing CLA on project of \"%s\"",
			orgInfo.OrgAlias,
		),
		emailtmpl.VerificationCode{
			Email:      req.Email,
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
