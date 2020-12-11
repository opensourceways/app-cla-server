package controllers

import (
	"fmt"
	"net/http"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type CorporationManagerController struct {
	baseController
}

func (this *CorporationManagerController) Prepare() {
	switch this.getRequestMethod() {
	case http.MethodPut:
		// add administrator
		this.apiPrepare(PermissionOwnerOfOrg)

	case http.MethodPatch:
		// reset password of manager
		this.apiPrepareForSettingPW()
	}
}

// @Title authenticate corporation manager
// @Description authenticate corporation manager
// @Param	body		body 	models.CorporationManagerAuthentication	true		"body for corporation manager info"
// @Success 201 {int} map
// @Failure util.ErrNoCLABindingDoc	"no cla binding applied to corporation"
// @router /auth [post]
func (this *CorporationManagerController) Auth() {
	doWhat := "authenticate as corp/employee manager"

	var info models.CorporationManagerAuthentication
	if err := this.fetchInputPayload(&info); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, doWhat)
		return
	}

	v, err := (&info).Authenticate()
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
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

	this.sendResponse(result, 0)
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

// @Title Put
// @Description add corporation administrator
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 202 {int} map
// @Failure util.ErrPDFHasNotUploaded
// @Failure util.ErrNumOfCorpManagersExceeded
// @router /:link_id/:email [put]
func (this *CorporationManagerController) Put() {
	doWhat := "add corp administrator"

	linkID := this.GetString(":link_id")
	corpEmail := this.GetString(":email")

	pl, err := this.tokenPayloadOfCodePlatform()
	if !pl.hasLink(linkID) {
		//TODO
	}

	uploaded, err := models.IsCorpSigningPDFUploaded(linkID, corpEmail)
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}
	if !uploaded {
		err = fmt.Errorf("pdf corporation signed has not been uploaded")
		this.sendFailedResponse(400, util.ErrPDFHasNotUploaded, err, doWhat)
		return
	}

	corpSigning, err := models.GetCorporationSigningBasicInfo(linkID, corpEmail)
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	added, err := models.CreateCorporationAdministrator(linkID, corpSigning.AdminName, corpEmail)
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	this.sendResponse("add manager successfully", 0)

	//TODO
	notifyCorpManagerWhenAdding(&models.OrgCLA{}, added)
}

// @Title Patch
// @Description reset password of corporation administrator
// @Success 204 {int} map
// @Failure util.ErrInvalidAccountOrPw
// @router / [patch]
func (this *CorporationManagerController) Patch() {
	doWhat := "reset password of corp's manager"

	pl, err := this.tokenPayloadOfCorpManager()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	var info models.CorporationManagerResetPassword
	if err := this.fetchInputPayload(&info); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, doWhat)
		return
	}

	if errCode, err := info.Validate(); err != nil {
		this.sendFailedResponse(400, errCode, err, doWhat)
		return
	}

	if err := (&info).Reset(pl.OrgCLAID, pl.Email); err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	this.sendResponse("reset password successfully", 0)
}

func (this *CorporationManagerController) apiPrepareForSettingPW() {
	if err := this.checkPathParameter(); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, "")
		this.StopRun()
	}

	if v := this.checkApiReqTokenForSettingPW(); v != nil {
		this.sendFailedResponse(v.statusCode, v.errCode, v.reason, "")
		this.StopRun()
	}
}

func (this *CorporationManagerController) checkApiReqTokenForSettingPW() *failedResult {
	token := this.apiReqHeader(headerToken)
	if token == "" {
		return &failedResult{
			statusCode: 401,
			errCode:    util.ErrMissingToken,
			reason:     fmt.Errorf("no token passed"),
		}
	}

	ac := &accessController{Payload: &acForCorpManagerPayload{}}
	if err := ac.ParseToken(token, conf.AppConfig.APITokenKey); err != nil {
		return &failedResult{
			statusCode: 401,
			errCode:    util.ErrUnknownToken,
			reason:     err,
		}
	}

	if err := ac.Verify([]string{PermissionCorporAdmin, PermissionEmployeeManager}); err != nil {
		return &failedResult{
			statusCode: 403,
			errCode:    util.ErrInvalidToken,
			reason:     err,
		}
	}

	this.Data[apiAccessController] = *ac
	return nil
}
