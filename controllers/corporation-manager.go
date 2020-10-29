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
		apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, &acForCodePlatformPayload{})

	case http.MethodPatch:
		// reset password of manager
		apiPrepare(&this.Controller, []string{PermissionCorporAdmin, PermissionEmployeeManager}, nil)
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
		dbmodels.CorporationManagerCheckResult
		Token    string `json:"token"`
		OrgCLAID string `json:"cla_org_id"`
	}

	result := make([]authInfo, 0, len(v))

	for orgCLAID, item := range v {
		token, err := this.newAccessToken(orgCLAID, item.Email, item.Role)
		if err != nil {
			continue
		}

		// should not expose email of corp manager
		item.Email = ""

		result = append(result, authInfo{
			CorporationManagerCheckResult: item,
			Token:                         token,
			OrgCLAID:                      orgCLAID,
		})
	}

	body = result
}

func (this *CorporationManagerController) newAccessToken(orgCLAID, email, role string) (string, error) {
	permission := ""
	switch role {
	case dbmodels.RoleAdmin:
		permission = PermissionCorporAdmin
	case dbmodels.RoleManager:
		permission = PermissionEmployeeManager
	}

	ac := &accessController{
		Expiry:     util.Expiry(conf.AppConfig.APITokenExpiry),
		Permission: permission,
		Payload: &accessControllerBasicPayload{
			User: corpManagerUser(orgCLAID, email),
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

	added, err := models.CreateCorporationAdministrator(orgCLAID, adminEmail)
	if err != nil {
		reason = err
		return
	}

	body = "add manager successfully"

	notifyCorpManagerWhenAdding(orgCLA, added)
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

	orgCLAID, corpEmail, err := parseCorpManagerUser(&this.Controller)
	if err != nil {
		reason = err
		errCode = util.ErrUnknownToken
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

	if err := (&info).Reset(orgCLAID, corpEmail); err != nil {
		reason = err
		return
	}

	body = "reset password successfully"
}
