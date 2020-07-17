package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/models"
)

type OrgRepoController struct {
	beego.Controller
}

// @Title Bind CLA to Org/Repo
// @Description bind cla
// @Param	body		body 	models.OrgRepo	true		"body for org-repo content"
// @Success 201 {int} models.OrgRepo
// @Failure 403 body is empty
// @router / [post]
func (this *OrgRepoController) Post() {
	var statusCode = 201
	var reason error

	defer func() {
		sendResponse(&this.Controller, statusCode, reason)
	}()

	var orgRepo models.OrgRepo

	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &orgRepo); err != nil {
		reason = err
		statusCode = 400
		return
	}

	cla := &models.CLA{ID: orgRepo.CLAID}

	if err := cla.Get(); err != nil {
		reason = fmt.Errorf("error finding the cla(id:%s), err: %v", cla.ID, err)
		statusCode = 400
		return
	}

	if cla.Language == "" {
		reason = fmt.Errorf("the language of cla(id:%s) is empty", cla.ID)
		statusCode = 500
		return
	}

	orgRepo.CLALanguage = cla.Language

	if err := (&orgRepo).Create(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	this.Data["json"] = orgRepo
}

// @Title Unbind CLA to Org/Repo
// @Description unbind cla
// @Param	uid		path 	string	true		"The uid of binding"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:uid [delete]
func (this *OrgRepoController) Delete() {
	var statusCode = 204
	var reason error

	defer func() {
		sendResponse(&this.Controller, statusCode, reason)
	}()

	uid := this.GetString(":uid")
	if uid == "" {
		reason = fmt.Errorf("missing binding id")
		statusCode = 400
		return
	}

	orgRepo := models.OrgRepo{ID: uid}

	if err := orgRepo.Delete(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	this.Data["json"] = "unbinding successfully"
}
