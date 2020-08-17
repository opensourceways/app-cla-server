package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/controllers/platforms"
	"github.com/zengchen1024/cla-server/models"
)

type CLAOrgController struct {
	beego.Controller
}

// @Title Bind CLA to Org/Repo
// @Description bind cla
// @Param	body		body 	models.CLAOrg	true		"body for org-repo content"
// @Success 201 {int} models.CLAOrg
// @Failure 403 body is empty
// @router / [post]
func (this *CLAOrgController) Post() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var claOrg models.CLAOrg

	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &claOrg); err != nil {
		reason = err
		statusCode = 400
		return
	}

	cla := &models.CLA{ID: claOrg.CLAID}

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

	if cla.ApplyTo == "" {
		reason = fmt.Errorf("the apply_to of cla(id:%s) is empty", cla.ID)
		statusCode = 500
		return
	}

	claOrg.CLALanguage = cla.Language
	claOrg.ApplyTo = cla.ApplyTo

	if err := (&claOrg).Create(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = claOrg
}

// @Title Unbind CLA from Org/Repo
// @Description unbind cla
// @Param	uid		path 	string	true		"The uid of binding"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:uid [delete]
func (this *CLAOrgController) Delete() {
	var statusCode = 204
	var reason error
	var body string

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	uid := this.GetString(":uid")
	if uid == "" {
		reason = fmt.Errorf("missing binding id")
		statusCode = 400
		return
	}

	claOrg := models.CLAOrg{ID: uid}

	if err := claOrg.Delete(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "unbinding successfully"
}

// @Title GetAll
// @Description get all bindings
// @Success 200 {object} models.CLAOrg
// @router / [get]
func (this *CLAOrgController) GetAll() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	h := parseHeader(&this.Controller)
	p, err := platforms.NewPlatform(h.accessToken, h.refreshToken, h.platform)
	if err != nil {
		reason = err
		statusCode = 400
		return
	}

	orgs, err := p.ListOrg()
	if err != nil {
		reason = fmt.Errorf("list org failed: %v", err)
		statusCode = 500
		return
	}

	opt := models.CLAOrgListOption{Org: map[string][]string{h.platform: orgs}}

	r, err := opt.List()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = r
}

// @Title Get signing page info
// @Description get signing page info
// @Success 200 {object} models.CLAOrg
// @router /signing-page [get]
func (this *CLAOrgController) GetSigningPageInfo() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	opt := models.CLAOrgListOption{
		Platform: this.GetString("platform"),
		OrgID:    this.GetString("org_id"),
		RepoID:   this.GetString("repo_id"),
		ApplyTo:  this.GetString("apply_to"),
	}

	r, err := opt.List()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = r
}
