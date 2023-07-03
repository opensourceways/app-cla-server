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

// @Title authenticate corporation manager
// @Description authenticate corporation manager
// @Param	body		body 	models.CorporationManagerAuthentication	true		"body for corporation manager info"
// @Success 201 {int} controllers.corpAuthInfo
// @Failure util.ErrNoCLABindingDoc	"no cla binding applied to corporation"
// @router /auth [post]
func (this *CorporationManagerController) Auth() {
	action := "authenticate as corp/employee manager"

	var info models.CorporationManagerAuthentication
	if fr := this.fetchInputPayload(&info); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if merr := info.Validate(); merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	orgInfo, merr := models.GetOrgOfLink(info.LinkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)

		return
	}

	v, merr := models.CorpManagerLogin(&info)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	if err := this.genToken(info.LinkID, &v); err != nil {
		this.sendFailedResponse(500, errSystemError, err, action)

		return
	}

	this.sendSuccessResp([]corpAuthInfo{
		{
			Role:             v.Role,
			OrgRepo:          orgInfo.OrgRepo,
			InitialPWChanged: v.InitialPWChanged,
		},
	})
}

func (this *CorporationManagerController) genToken(linkID string, info *models.CorpManagerLoginInfo) error {
	permission := ""
	switch info.Role {
	case dbmodels.RoleAdmin:
		permission = PermissionCorpAdmin
	case dbmodels.RoleManager:
		permission = PermissionEmployeeManager
	}

	token, err := this.newApiToken(
		permission,
		&acForCorpManagerPayload{
			Corp:      info.CorpName,
			Email:     info.Email,
			LinkID:    linkID,
			SigningId: info.SigningId,
		},
	)
	if err == nil {
		this.setToken(token)
	}

	return err
}

type acForCorpManagerPayload struct {
	Corp      string `json:"corp"`
	Email     string `json:"email"`
	LinkID    string `json:"link_id"`
	SigningId string `json:"csid"`
}
