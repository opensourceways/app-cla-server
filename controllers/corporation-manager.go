package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

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
		CLAOrgID string `json:"cla_org_id"`
	}

	result := make([]authInfo, 0, len(v))

	for claOrgID, items := range v {
		for _, item := range items {
			user := corpManagerUser(claOrgID, item.Email)
			token, err := newAccessToken(user, corporRoleToPermission(item.Role))
			if err != nil {
				continue
			}

			// should not expose email of corp manager
			item.Email = ""

			result = append(result, authInfo{
				CorporationManagerCheckResult: item,
				Token:                         token,
				CLAOrgID:                      claOrgID,
			})
		}
	}

	body = result
}

// @Title Put
// @Description add corporation administrator
// @Param	:cla_org_id	path 	string					true		"cla org id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 202 {int} map
// @Failure util.ErrPDFHasNotUploaded
// @Failure util.ErrNumOfCorpManagersExceeded
// @router /:cla_org_id/:email [put]
func (this *CorporationManagerController) Put() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "add corp administrator")
	}()

	if err := checkAPIStringParameter(&this.Controller, []string{":cla_org_id", ":email"}); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}
	claOrgID := this.GetString(":cla_org_id")
	adminEmail := this.GetString(":email")

	var claOrg *models.CLAOrg
	claOrg, statusCode, errCode, reason = canOwnerOfOrgAccessCLA(&this.Controller, claOrgID)
	if reason != nil {
		return
	}

	info, err := models.CheckCorporationSigning(claOrgID, adminEmail)
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

	added, err := models.CreateCorporationAdministrator(claOrgID, adminEmail)
	if err != nil {
		reason = err
		return
	}

	body = "add manager successfully"

	notifyCorpManagerWhenAdding(claOrg.OrgEmail, "Corporation Administrator", added)
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

	claOrgID, corpEmail, err := parseCorpManagerUser(&this.Controller)
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

	if err := (&info).Reset(claOrgID, corpEmail); err != nil {
		reason = err
		return
	}

	body = "reset password successfully"
}
