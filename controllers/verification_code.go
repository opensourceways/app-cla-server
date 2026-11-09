package controllers

import (
	"errors"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
)

func (ctl *baseController) sendVerificationCodeWhenSigning(
	linkID string, f func(string) (string, models.IModelError),
) {
	action := "send verification code when signing"

	var req verificationCodeRequest
	if fr := ctl.fetchInputPayload(&req); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if err := req.validate(); err != nil {
		ctl.sendFailedResultAsResp(
			newFailedApiResult(400, errParsingApiBody, err),
			action,
		)
		return
	}

	orgInfo, merr := models.GetLink(linkID)
	if merr != nil {
		ctl.sendFailedResponse(0, "", merr, action)
		return
	}

	code, err := f(req.Email)
	if err != nil {
		ctl.sendModelErrorAsResp(err, action)
		return
	}

	ctl.sendSuccessResp(action, "successfully")

	sendEmailToIndividual(
		req.Email, &orgInfo,
		&emailtmpl.VerificationCode{
			Lang:        req.Language,
			Email:       req.Email,
			Org:         orgInfo.OrgAlias,
			Code:        code,
			ProjectURL:  orgInfo.ProjectURL,
		},
	)
}

type verificationCodeRequest struct {
	Email    string `json:"email" required:"true"`
	Language string `json:"language"`
}

func (req *verificationCodeRequest) validate() error {
	if req.Email == "" {
		return errors.New("missing email")
	}

	return nil
}
