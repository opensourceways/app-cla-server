package controllers

import (
	"encoding/json"
	"errors"

	"github.com/opensourceways/app-cla-server/models"
)

const (
	PermissionCorpAdmin       = "corporation administrator"
	PermissionOwnerOfOrg      = "owner of org"
	PermissionEmployeeManager = "employee manager"
)

type privacyConsentCheck interface {
	checkPrivacyConsent(string) error
}

type accessController struct {
	RemoteAddr string      `json:"remote_addr"`
	Permission string      `json:"permission"`
	Payload    interface{} `json:"payload"`
}

func (ctl *accessController) getUser() string {
	pl, ok := ctl.Payload.(*acForCorpManagerPayload)
	if !ok {
		return ""
	}

	if ctl.Permission == PermissionOwnerOfOrg {
		return pl.UserId
	}

	return pl.LinkID + "/" + pl.UserId
}

func (ctl *accessController) verify(permission []string, addr string) error {
	if ctl.RemoteAddr != addr {
		return errors.New("unmatched remote address")
	}

	for _, p := range permission {
		if p == ctl.Permission {
			return nil
		}
	}

	return errors.New("not allowed permission")
}

func (ctl *accessController) checkPrivacyConsent() error {
	if v, ok := ctl.Payload.(privacyConsentCheck); ok {
		return v.checkPrivacyConsent(privacyVersion)
	}

	return nil
}

func (ctl *baseController) apiPrepare(permission string) {
	if permission != "" {
		ac := ctl.newAccessController(permission)

		ctl.apiPrepareWithAC(&ac, []string{permission})
	} else {
		ctl.apiPrepareWithAC(nil, nil)
	}
}

func (ctl *baseController) apiPrepareWithAC(ac *accessController, permission []string) {
	if fr := ctl.checkPathParameter(); fr != nil {
		ctl.sendFailedResultAsResp(fr, "")
		ctl.StopRun()
	}

	if ac != nil && len(permission) != 0 {
		if fr := ctl.checkApiReqToken(ac, permission); fr != nil {
			ctl.sendFailedResultAsResp(fr, "")
			ctl.StopRun()
		}

		ctl.Data[apiAccessController] = *ac
	}
}

func (ctl *baseController) newAccessController(permission string) accessController {
	var acp interface{}

	switch permission {
	case PermissionOwnerOfOrg:
		acp = &acForCorpManagerPayload{}
	case PermissionCorpAdmin:
		acp = &acForCorpManagerPayload{}
	case PermissionEmployeeManager:
		acp = &acForCorpManagerPayload{}
	}

	return accessController{Payload: acp}
}

func (ctl *baseController) checkApiReqToken(ac *accessController, permission []string) *failedApiResult {
	token, fr := ctl.getToken()
	if fr != nil {
		return fr
	}

	newToken, v, err := models.ValidateAndRefreshAccessToken(token)

	if err != nil {
		if err.IsErrorOf(models.ErrInvalidToken) {
			return newFailedApiResult(401, errUnknownToken, err)
		}

		return newFailedApiResult(500, errSystemError, err)
	}

	ctl.setToken(newToken)

	if err := json.Unmarshal(v, ac); err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}

	addr, fr := ctl.getRemoteAddr()
	if fr != nil {
		return fr
	}

	if err := ac.verify(permission, addr); err != nil {
		return newFailedApiResult(403, errUnauthorizedToken, err)
	}

	if err := ac.checkPrivacyConsent(); err != nil {
		return newFailedApiResult(401, models.ErrPrivacyConsentInvalid, err)
	}

	return nil
}

func (ctl *baseController) getAccessController() (*accessController, *failedApiResult) {
	ac, ok := ctl.Data[apiAccessController]
	if !ok {
		return nil, newFailedApiResult(500, errSystemError, errors.New("no access controller"))
	}

	if v, ok := ac.(accessController); ok {
		return &v, nil
	}

	return nil, newFailedApiResult(500, errSystemError, errors.New("can't convert to access controller instance"))
}
