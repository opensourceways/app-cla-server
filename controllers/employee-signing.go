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
