package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/models"
)

type EmployeeManagerController struct {
	beego.Controller
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
