package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
)

// @Title authenticate corporation manager
// @Description authenticate corporation manager
// @Param	body		body 	models.CorporationManagerAuthentication	true		"body for corporation manager info"
// @Success 201 {int} map
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
	option, mErr := models.InitializeLoginMissOption(info.LinkID, info.User)
	if mErr != nil {
		this.sendModelErrorAsResp(mErr, action)
	}

	if option.IsLocked() {
		this.sendFailedResponse(400, errAccountLocked, fmt.Errorf("account locked with more error login"), action)
		return
	}
	option.ResetLockExpiredLoginNum()

	v, mErr := (&info).Authenticate()
	if mErr != nil {
		this.sendModelErrorAsResp(mErr, action)
		option.DoLoginMiss()
		return
	}
	if len(v) == 0 {
		this.sendFailedResponse(400, errWrongIDOrPassword, fmt.Errorf("wrong id or pw"), action)
		option.DoLoginMiss()
		return
	}

	type authInfo struct {
		models.OrgRepo

		Role             string `json:"role"`
		Token            string `json:"token"`
		InitialPWChanged bool   `json:"initial_pw_changed"`
	}

	result := make([]authInfo, 0, len(v))

	for linkID, item := range v {
		token, err := this.newAccessToken(linkID, &item)
		if err != nil {
			continue
		}

		result = append(result, authInfo{
			OrgRepo:          item.OrgRepo,
			Role:             item.Role,
			Token:            token,
			InitialPWChanged: item.InitialPWChanged,
		})
	}

	this.sendSuccessResp(result)
	option.DoLoginSuccess()
}

func (this *CorporationManagerController) newAccessToken(linkID string, info *dbmodels.CorporationManagerCheckResult) (string, error) {
	permission := ""
	switch info.Role {
	case dbmodels.RoleAdmin:
		permission = PermissionCorpAdmin
	case dbmodels.RoleManager:
		permission = PermissionEmployeeManager
	}

	return this.newApiToken(
		permission,
		&acForCorpManagerPayload{
			Corp:    info.Corp,
			Name:    info.Name,
			Email:   info.Email,
			LinkID:  linkID,
			OrgInfo: info.OrgInfo,
		},
	)
}

type acForCorpManagerPayload struct {
	Corp   string `json:"corp"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	LinkID string `json:"link_id"`

	models.OrgInfo
}
