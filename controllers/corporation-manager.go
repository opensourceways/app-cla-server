package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
)

type CorporationManagerController struct {
	beego.Controller
}

func (this *CorporationManagerController) Prepare() {
	if getRouterPattern(&this.Controller) == "/v1/corporation-manager/:cla_org_id/:email" {
		if this.Ctx.Request.Method == http.MethodPut {
			apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, nil)
		} else {
			apiPrepare(&this.Controller, []string{PermissionCorporAdmin, PermissionEmployeeManager}, nil)
		}
	}
}

// @Title Put
// @Description add corporation administrator
// @Param	:cla_org_id	path 	string					true		"cla org id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 202 {int} map
// @router /:cla_org_id/:email [put]
func (this *CorporationManagerController) Put() {
	var statusCode = 0
	var errCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body)
	}()

	if err := checkAPIStringParameter(&this.Controller, []string{":cla_org_id", ":email"}); err != nil {
		reason = err
		errCode = ErrInvalidParameter
		statusCode = 400
		return
	}
	claOrgID := this.GetString(":cla_org_id")
	email := this.GetString(":email")

	info, err := models.CheckCorporationSigning(claOrgID, email)
	if err != nil {
		reason = fmt.Errorf("Failed to add corp administrator, err: %s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	if !info.PDFUploaded {
		reason = fmt.Errorf("Failed to add corp administrator, err: pdf corporation signed has not been uploaded")
		errCode = ErrPDFHasNotUploaded
		statusCode = 400
		return
	}

	if info.AdminAdded {
		// TODO: send email failed
	}

	err = models.CreateCorporationAdministrator(claOrgID, email)
	if err != nil {
		reason = fmt.Errorf("Failed to add corp administrator, err: %s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	body = "add manager successfully"
}

// @Title authenticate corporation manager
// @Description authenticate corporation manager
// @Param	body		body 	models.CorporationManagerAuthentication	true		"body for corporation manager info"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router /auth [post]
func (this *CorporationManagerController) Auth() {
	var statusCode = 201
	var errCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body)
	}()

	var info models.CorporationManagerAuthentication
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = ErrInvalidParameter
		statusCode = 400
		return
	}

	v, err := (&info).Authenticate()
	if err != nil {
		reason = fmt.Errorf("Failed to authenticate as corp manager, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
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
			token, err := newAccessToken(item.Email, corporRoleToPermission(item.Role))
			if err != nil {
				continue
			}

			result = append(result, authInfo{
				CorporationManagerCheckResult: item,
				Token:                         token,
				CLAOrgID:                      claOrgID,
			})
		}
	}

	body = result
}

// @Title Patch
// @Description reset password of corporation administrator
// @Param	:cla_org_id	path 	string					true		"cla org id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 204 {int} map
// @router /:cla_org_id/:email [patch]
func (this *CorporationManagerController) Patch() {
	var statusCode = 0
	var errCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body)
	}()

	if err := checkAPIStringParameter(&this.Controller, []string{":cla_org_id", ":email"}); err != nil {
		reason = err
		errCode = ErrInvalidParameter
		statusCode = 400
		return
	}

	var info models.CorporationManagerResetPassword
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = ErrInvalidParameter
		statusCode = 400
		return
	}

	if err := (&info).Reset(this.GetString(":cla_org_id"), this.GetString(":email")); err != nil {
		reason = fmt.Errorf("Failed to reset password of admin, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	body = "reset password successfully"
}
