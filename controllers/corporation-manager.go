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
	if getRouterPattern(&this.Controller) == "/v1/corporation-manager/:cla_org_id/:email" {
		if this.Ctx.Request.Method == http.MethodPut {
			// add administrator
			apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, nil)
		} else {
			// reset password of manager
			apiPrepare(&this.Controller, []string{PermissionCorporAdmin, PermissionEmployeeManager}, nil)
		}
	}
}

// @Title authenticate corporation manager
// @Description authenticate corporation manager
// @Param	body		body 	models.CorporationManagerAuthentication	true		"body for corporation manager info"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router /auth [post]
func (this *CorporationManagerController) Auth() {
	var statusCode = 201
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "authenticate as corp manager")
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
	email := this.GetString(":email")

	info, err := models.CheckCorporationSigning(claOrgID, email)
	if err != nil {
		reason = err
		statusCode, errCode = convertDBError(err)
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
	}

	err = models.CreateCorporationAdministrator(claOrgID, email)
	if err != nil {
		reason = err
		statusCode, errCode = convertDBError(err)
		return
	}

	body = "add manager successfully"
}

// @Title Patch
// @Description reset password of corporation administrator
// @Param	:cla_org_id	path 	string					true		"cla org id"
// @Success 204 {int} map
// @router /:cla_org_id [patch]
func (this *CorporationManagerController) Patch() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "reset password of corp's manager")
	}()

	claOrgID, err := fetchStringParameter(&this.Controller, ":cla_org_id")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	corpEmail, err := getApiAccessUser(&this.Controller)
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	var info models.CorporationManagerResetPassword
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	if err := (&info).Reset(claOrgID, corpEmail); err != nil {
		reason = err
		statusCode, errCode = convertDBError(err)
		return
	}

	body = "reset password successfully"
}
