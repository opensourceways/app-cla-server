package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/huaweicloud/golangsdk"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/models"
)

type CorporationManagerController struct {
	beego.Controller
}

func (this *CorporationManagerController) Prepare() {
	method := this.Ctx.Request.Method

	if method == http.MethodPost {
		if getRouterPattern(&this.Controller) == "/v1/corporation-manager/auth" {
			return
		}
		apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, nil)
	} else {
		apiPrepare(&this.Controller, []string{PermissionCorporAdmin, PermissionEmployeeManager}, nil)
	}
}

// @Title add corporation manager
// @Description add corporation manager
// @Param	body		body 	models.CorporationManagerCreateOption	true		"body for corporation manager"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router / [post]
func (this *CorporationManagerController) Post() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.CorporationManagerCreateOption
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if err := (&info).Create(); err != nil {
		reason = err
		statusCode = 500
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
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.CorporationManagerAuthentication
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}

	v, err := (&info).Authenticate()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}
	if len(v) == 0 {
		reason = fmt.Errorf("the user or password is not correct")
		statusCode = 500
		return
	}

	result := make([]map[string]interface{}, 0, len(v))
	for _, item := range v {
		ac := &accessController{
			User:       item.Email,
			Permission: corporRoleToPermission(item.Role),
			Expiry:     conf.AppConfig.APITokenExpiry,
		}

		token, err := ac.CreateToken(conf.AppConfig.APITokenKey)
		if err != nil {
			continue
		}

		m, err := golangsdk.BuildRequestBody(item, "")
		if err != nil {
			continue
		}

		m["token"] = token
		result = append(result, m)
	}
	body = result
}

// @Title Reset password
// @Description reset password
// @Param	body		body 	models.CorporationManagerResetPassword	true		"body for corporation manager info"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router / [put]
func (this *CorporationManagerController) Update() {
	var statusCode = 202
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.CorporationManagerResetPassword
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if err := (&info).Reset(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "reset password successfully"
}
