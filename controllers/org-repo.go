package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/models"
)

type OrgRepoController struct {
	beego.Controller
}

// @Title Bind CLA to Org/Repo
// @Description bind cla
// @Param	body		body 	models.OrgRepo	true		"body for org-repo content"
// @Success 200 {int} models.OrgRepo
// @Failure 403 body is empty
// @router / [post]
func (this *OrgRepoController) Post() {
	defer func() {
		this.ServeJSON()
	}()

	var orgRepo models.OrgRepo

	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &orgRepo); err != nil {
		this.Data["json"] = err.Error()
		return
	}

	r, err := orgRepo.Create()
	if err != nil {
		this.Data["json"] = err.Error()
		return
	}

	this.Data["json"] = r
}
