package controllers

import (
	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

// @Title authenticate corporation manager
// @Description authenticate corporation manager
// @Param	body		body 	models.CorporationManagerAuthentication	true		"body for corporation manager info"
// @Success 201 {int} map
// @Failure util.ErrNoCLABindingDoc	"no cla binding applied to corporation"
// @router /auth [post]
func (this *CorporationManagerController) Auth() {
	action := "authenticate as corp/employee manager"
	sendResp := this.newFuncForSendingFailedResp(action)

	var info models.CorporationManagerAuthentication
	if fr := this.fetchInputPayload(&info); fr != nil {
		sendResp(fr)
		return
	}

	v, err := (&info).Authenticate()
	if err != nil {
		sendResp(convertDBError1(err))
		return
	}

	type authInfo struct {
		Role             string `json:"role"`
		Platform         string `json:"platform"`
		OrgID            string `json:"org_id"`
		RepoID           string `json:"repo_id"`
		Token            string `json:"token"`
		InitialPWChanged bool   `json:"initial_pw_changed"`
	}

	result := make([]authInfo, 0, len(v))

	for orgCLAID, item := range v {
		token, err := this.newAccessToken(orgCLAID, &item)
		if err != nil {
			continue
		}

		result = append(result, authInfo{
			Role:             item.Role,
			Platform:         item.Platform,
			OrgID:            item.OrgID,
			RepoID:           item.RepoID,
			Token:            token,
			InitialPWChanged: item.InitialPWChanged,
		})
	}

	this.sendSuccessResp(result)
}

func (this *CorporationManagerController) newAccessToken(orgCLAID string, info *dbmodels.CorporationManagerCheckResult) (string, error) {
	permission := ""
	switch info.Role {
	case dbmodels.RoleAdmin:
		permission = PermissionCorporAdmin
	case dbmodels.RoleManager:
		permission = PermissionEmployeeManager
	}

	ac := &accessController{
		Expiry:     util.Expiry(conf.AppConfig.APITokenExpiry),
		Permission: permission,
		Payload: &acForCorpManagerPayload{
			Name:     info.Name,
			Email:    info.Email,
			OrgCLAID: orgCLAID,
		},
	}

	return ac.NewToken(conf.AppConfig.APITokenKey)
}
