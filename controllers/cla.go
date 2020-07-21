package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/models"
)

type CLAController struct {
	beego.Controller
}

// @Title CreateCLA
// @Description create cla
// @Param	body		body 	models.CLA	true		"body for cla content"
// @Success 201 {int} models.CLA
// @Failure 403 body is empty
// @router / [post]
func (this *CLAController) Post() {
	var statusCode = 201
	var reason error

	defer func() {
		sendResponse(&this.Controller, statusCode, reason)
	}()

	var cla models.CLA
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &cla); err != nil {
		reason = err
		statusCode = 400
		return
	}

	submitter := getHeader(&this.Controller, headerUser)
	cla.Submitter = submitter

	if err := (&cla).Create(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	this.Data["json"] = cla
}

// @Title GetAllCLA
// @Description get all clas
// @Success 200 {object} models.CLA
// @router / [get]
func (this *CLAController) GetAll() {
	var statusCode = 200
	var reason error

	defer func() {
		sendResponse(&this.Controller, statusCode, reason)
	}()

	clas := models.CLAs{BelongTo: []string{getHeader(&this.Controller, headerUser)}}

	r, err := clas.Get()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	this.Data["json"] = r
}
