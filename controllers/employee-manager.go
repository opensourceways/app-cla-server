package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
)

type EmployeeManagerController struct {
	beego.Controller
}

func (this *EmployeeManagerController) Prepare() {
	apiPrepare(&this.Controller, []string{PermissionCorporAdmin})
}

// @Title add employee manager
// @Description add employee manager
// @Param	body		body 	models.EmployeeManagerCreateOption	true		"body for employee manager"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router / [post]
func (this *EmployeeManagerController) Post() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.EmployeeManagerCreateOption
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if err := (&info).Validate(); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if err := (&info).Create(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "add employee manager successfully"
}

// @Title GetAll
// @Description get all employee managers
// @Success 200 {object} dbmodels.CorporationManagerListResult
// @router / [get]
func (this *EmployeeManagerController) GetAll() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	opt := models.CorporationManagerListOption{
		CLAOrgID: this.GetString("cla_org_id"),
		Email:    this.GetString("email"),
		Role:     dbmodels.RoleManager,
	}

	r, err := opt.List()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = r
}

// @Title Delete
// @Description delete employee manager
// @Param	body		body 	models.EmployeeManagerCreateOption	true		"body for employee manager"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router / [delete]
func (this *EmployeeManagerController) Delete() {
	var statusCode = 204
	var reason error
	var body string

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.EmployeeManagerCreateOption
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if err := (&info).Validate(); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if err := (&info).Delete(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "delete cla successfully"
}
