package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type CorporationManagerController struct {
	beego.Controller
}

func (this *CorporationManagerController) Prepare() {
	switch getRequestMethod(&this.Controller) {
	case http.MethodPut:
		// add administrator
		apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg})

	case http.MethodPatch:
		// reset password of manager
		apiPrepare(&this.Controller, []string{PermissionCorporAdmin, PermissionEmployeeManager})
	}
}

// @Title authenticate corporation manager
// @Description authenticate corporation manager
// @Param	body		body 	models.CorporationManagerAuthentication	true		"body for corporation manager info"
// @Success 201 {int} map
// @Failure util.ErrNoCLABindingDoc	"no cla binding applied to corporation"
// @router /auth [post]
func (this *CorporationManagerController) Auth() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "authenticate as corp/employee manager")
	}()

	var info models.CorporationManagerAuthentication
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	v, err := (&info).Authenticate()
	if err != nil {
		reason = err
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

	for _, item := range v {
		token, err := this.newAccessToken(&item)
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

	body = result
}

func (this *CorporationManagerController) newAccessToken(info *dbmodels.CorporationManagerCheckResult) (string, error) {
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
			OrgID:    info.OrgID,
			RepoID:   info.RepoID,
			Platform: info.Platform,
			OrgEmail: info.OrgEmail,
			OrgAlias: info.OrgAlias,
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
// @router /:org_cla_id/:email [put]
func (this *CorporationManagerController) Put() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "add corp administrator")
	}()

	if err := checkAPIStringParameter(&this.Controller, []string{":org_cla_id", ":email"}); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}
	orgCLAID := this.GetString(":org_cla_id")
	adminEmail := this.GetString(":email")

	var orgCLA *models.OrgCLA
	orgCLA, statusCode, errCode, reason = canAccessOrgCLA(&this.Controller, orgCLAID)
	if reason != nil {
		return
	}

	info, err := models.CheckCorporationSigning(orgCLAID, adminEmail)
	if err != nil {
		reason = err
		return
	}

	if !info.PDFUploaded {
		reason = fmt.Errorf("pdf corporation signed has not been uploaded")
		errCode = util.ErrPDFHasNotUploaded
		statusCode = 400
		return
	}

	if info.AdminAdded {
		// TODO: send email failed
		return
	}

	orgRepo := buildOrgRepo(orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID)
	added, err := models.CreateCorporationAdministrator(&orgRepo, info.AdminName, adminEmail)
	if err != nil {
		reason = err
		return
	}

	body = "add manager successfully"

	notifyCorpManagerWhenAdding(orgCLA.OrgAlias, projectURL(orgCLA), orgCLA.OrgEmail, added)
}

// @Title Patch
// @Description reset password of corporation administrator
// @Success 204 {int} map
// @Failure util.ErrInvalidAccountOrPw
// @router / [patch]
func (this *CorporationManagerController) Patch() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "reset password of corp's manager")
	}()

	var ac *acForCorpManagerPayload
	ac, errCode, reason = getACOfCorpManager(&this.Controller)
	if reason != nil {
		statusCode = 401
		return
	}

	var info models.CorporationManagerResetPassword
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	if errCode, reason = info.Validate(); reason != nil {
		statusCode = 400
		return
	}

	orgRepo := buildOrgRepo(ac.Platform, ac.OrgID, ac.RepoID)
	if err := (&info).Reset(&orgRepo, ac.Email); err != nil {
		reason = err
		return
	}

	body = "reset password successfully"
}
