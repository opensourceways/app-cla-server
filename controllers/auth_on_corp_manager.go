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

	v, merr := (&info).Authenticate()
	if merr != nil {
		sendResp(parseModelError(merr))
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
}

func (this *CorporationManagerController) newAccessToken(linkID string, info *dbmodels.CorporationManagerCheckResult) (string, error) {
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
			OrgCLAID: linkID,
			OrgInfo:  info.OrgInfo,
		},
	}

	return ac.NewToken(conf.AppConfig.APITokenKey)
}

type acForCorpManagerPayload struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	OrgCLAID string `json:"link_id"`

	models.OrgInfo
}

func (this *acForCorpManagerPayload) hasEmployee(email string) bool {
	return util.EmailSuffix(this.Email) == util.EmailSuffix(email)
}
