package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/models"
)

type EmployeeSigningController struct {
	beego.Controller
}

// @Title Employee signing
// @Description sign as employee
// @Param	body		body 	models.EmployeeSigning	true		"body for employee signing"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router / [post]
func (this *EmployeeSigningController) Post() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.EmployeeSigning
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

	body = "sign successfully"
}

// @Title GetAll
// @Description get all the employees
// @Success 200 {object} models.EmployeeSigning
// @router / [get]
func (this *EmployeeSigningController) GetAll() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	opt := models.EmployeeSigningListOption{
		Platform:         this.GetString("platform"),
		OrgID:            this.GetString("org_id"),
		RepoID:           this.GetString("repo_id"),
		CLALanguage:      this.GetString("cla_language"),
		CorporationEmail: this.GetString("corporation_email"),
	}

	r, err := opt.List()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = r
}

// @Title Enable employee signing
// @Description enable employee
// @Param	body		body 	models.EmployeeSigning	true		"body for employee signing"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router / [put]
func (this *EmployeeSigningController) Update() {
	var statusCode = 202
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.EmployeeSigningUdateInfo
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if err := (&info).Update(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "enabled employee successfully"
}
