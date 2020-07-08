package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla/models"
)

type CLAController struct {
	beego.Controller
}

// @Title CreateCLA
// @Description create cla
// @Param	body		body 	models.CLA	true		"body for cla content"
// @Success 200 {int} models.CLA
// @Failure 403 body is empty
// @router / [post]
func (this *CLAController) Post() {
	defer func() {
		this.ServeJSON()
	}()

	var cla models.CLA
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &cla); err != nil {
		this.Data["json"] = err.Error()
		return
	}

	cla1, err := cla.Create()
	if err != nil {
		this.Data["json"] = err.Error()
		return
	}

	this.Data["json"] = cla1
}
