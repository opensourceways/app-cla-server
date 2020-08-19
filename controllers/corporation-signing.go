package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/models"
)

type CorporationSigningController struct {
	beego.Controller
}

// @Title Corporation signing
// @Description sign as corporation
// @Param	body		body 	models.CorporationSigning	true		"body for corporation signing"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router / [post]
func (this *CorporationSigningController) Post() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.CorporationSigning
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
// @Description get all the corporations which have signed to a org
// @Success 200 {object} dbmodels.CorporationSigningInfo
// @router / [get]
func (this *CorporationSigningController) GetAll() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	opt := models.CorporationSigningListOption{
		Platform:    this.GetString("platform"),
		OrgID:       this.GetString("org_id"),
		RepoID:      this.GetString("repo_id"),
		CLALanguage: this.GetString("cla_language"),
	}

	r, err := opt.List()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = r
}

// @Title Enable corporation signing
// @Description enable corporation
// @Param	body		body 	models.CorporationSigning	true		"body for corporation signing"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router / [put]
func (this *CorporationSigningController) Update() {
	var statusCode = 202
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.CorporationSigningUdateInfo
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

	body = "enabled corporation successfully"
}
