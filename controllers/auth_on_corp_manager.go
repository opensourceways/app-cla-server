package controllers

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
)

type corpAuthInfo struct {
	models.OrgRepo

	Role             string `json:"role"`
	InitialPWChanged bool   `json:"initial_pw_changed"`
}

// @Title logout
// @Description corporation manager logout
// @Tags CorpManager
// @Accept json
// @Success 202 {int} controllers.corpAuthInfo
// @Failure util.ErrNoCLABindingDoc	"no cla binding applied to corporation"
// @router /auth [put]
func (ctl *CorporationManagerController) Logout() {
	action := "corp manager logout"
	sendResp := ctl.newFuncForSendingFailedResp(action)

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	models.CorpManagerLogout(pl.UserId)

	ctl.logout()

	ctl.sendSuccessResp(action + " successfully")
}

// @Title authenticate corporation manager
// @Description authenticate corporation manager
// @Tags CorpManager
// @Accept json
// @Param  body  body  models.CorporationManagerAuthentication  true  "body for corporation manager info"
// @Success 201 {int} controllers.corpAuthInfo
// @Failure util.ErrNoCLABindingDoc	"no cla binding applied to corporation"
// @router /auth [post]
func (ctl *CorporationManagerController) Auth() {
	action := "authenticate as corp/employee manager"

	var info models.CorporationManagerAuthentication
	if fr := ctl.fetchInputPayload(&info); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if merr := info.Validate(); merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
		return
	}

	orgInfo, merr := models.GetLink(info.LinkID)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)

		return
	}

	v, merr := models.CorpManagerLogin(&info)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
		return
	}

	if err := ctl.genToken(info.LinkID, &v); err != nil {
		ctl.sendFailedResponse(500, errSystemError, err, action)

		return
	}

	ctl.sendSuccessResp([]corpAuthInfo{
		{
			Role:             v.Role,
			OrgRepo:          orgInfo.OrgRepo,
			InitialPWChanged: v.InitialPWChanged,
		},
	})
}

func (ctl *CorporationManagerController) genToken(linkID string, info *models.CorpManagerLoginInfo) error {
	permission := ""
	switch info.Role {
	case dbmodels.RoleAdmin:
		permission = PermissionCorpAdmin
	case dbmodels.RoleManager:
		permission = PermissionEmployeeManager
	}

	token, err := ctl.newApiToken(
		permission,
		&acForCorpManagerPayload{
			Corp:      info.CorpName,
			Email:     info.Email,
			UserId:    info.UserId,
			LinkID:    linkID,
			SigningId: info.SigningId,
		},
	)
	if err == nil {
		ctl.setToken(token)
	}

	return err
}

type acForCorpManagerPayload struct {
	Corp      string `json:"corp"`
	Email     string `json:"email"`
	UserId    string `json:"user_id"`
	LinkID    string `json:"link_id"`
	SigningId string `json:"csid"`
}
