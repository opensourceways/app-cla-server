package controllers

import (
	"errors"

	"github.com/opensourceways/app-cla-server/models"
)

type corpAuthFailure struct {
	errMsg
	RetryNum int `json:"retry_num"`
}

// @Title Logout
// @Description corporation manager logout
// @Tags CorpManager
// @Accept json
// @Success 202 {object} controllers.respData
// @Failure util.ErrNoCLABindingDoc	"no cla binding applied to corporation"
// @router /auth [put]
func (ctl *CorporationManagerController) Logout() {
	action := "corp admin or employee manager logouts"

	ctl.logout()

	ctl.sendSuccessResp(action, "successfully")
}

// @Title Login
// @Description corporation manager login
// @Tags CorpManager
// @Accept json
// @Param  body  body  models.CorporationManagerLoginInfo  true  "body for corporation manager info"
// @Success 201
// @Failure util.ErrNoCLABindingDoc	"no cla binding applied to corporation"
// @router /auth [post]
func (ctl *CorporationManagerController) Login() {
	action := "corp admin or employee manager logins"

	var info models.CorporationManagerLoginInfo
	if fr := ctl.fetchInputPayload(&info); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if merr := info.Validate(); merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
		return
	}

	v, merr := models.CorpManagerLogin(&info)
	if merr != nil {
		if merr.IsErrorOf(models.ErrWrongIDOrPassword) {
			body := corpAuthFailure{
				RetryNum: v.RetryNum,
			}
			body.ErrCode = merr.ErrCode()
			body.ErrMsg = merr.Error()

			ctl.sendResponse(action, body, 400)

		} else if merr.IsErrorOf(models.ErrPrivacyConsentInvalid) {
			body := errMsg{
				ErrCode: merr.ErrCode(),
				ErrMsg:  merr.Error(),
			}

			ctl.sendResponse(action, body, 401)
		} else {
			ctl.sendModelErrorAsResp(merr, action)
		}
		return
	}

	if err := ctl.genToken(info.LinkID, &v); err != nil {
		ctl.sendFailedResponse(500, errSystemError, err, action)

		return
	}

	ctl.sendSuccessResp(action, "successfully")

	ctl.addOperationLog(v.UserId+" / "+v.Role, action, 0)
}

func (ctl *CorporationManagerController) genToken(linkID string, info *models.CorpManagerLoginInfo) error {
	permission := ""
	switch info.Role {
	case models.RoleAdmin:
		permission = PermissionCorpAdmin
	case models.RoleManager:
		permission = PermissionEmployeeManager
	case "":
		permission = PermissionOwnerOfOrg
	}

	token, err := ctl.newApiToken(
		permission,
		&acForCorpManagerPayload{
			Email:          info.Email,
			UserId:         info.UserId,
			LinkID:         linkID,
			SigningId:      info.SigningId,
			PrivacyVersion: info.PrivacyVersion,
		},
	)
	if err == nil {
		ctl.setToken(token)
	}

	return err
}

type acForCorpManagerPayload struct {
	Email          string `json:"email"`
	UserId         string `json:"user_id"`
	LinkID         string `json:"link_id"`
	SigningId      string `json:"csid"`
	PrivacyVersion string `json:"privacy"`
}

func (pl *acForCorpManagerPayload) checkPrivacyConsent(v string) error {
	return checkPrivacyConsent(pl.PrivacyVersion, v)
}

func checkPrivacyConsent(current, expect string) error {
	if current == "" {
		return errors.New("no privacy info")
	}

	if current != expect {
		return errors.New("privacy is not latest")
	}

	return nil
}
