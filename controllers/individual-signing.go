package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/models"
)

type IndividualSigningController struct {
	beego.Controller
}

func (this *IndividualSigningController) Prepare() {
	if getRouterPattern(&this.Controller) == "/v1/individual-signing/:platform/:org/:repo" {
		return
	}

	apiPrepare(&this.Controller, []string{PermissionIndividualSigner}, nil)
}

// @Title Individual signing
// @Description sign as individual
// @Param	body		body 	models.IndividualSigning	true		"body for individual signing"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router /:cla_org_id [post]
func (this *IndividualSigningController) Post() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	claOrgID := this.GetString(":cla_org_id")
	if claOrgID == "" {
		reason = fmt.Errorf("missing :cla_org_id")
		statusCode = 400
		return
	}

	var info models.IndividualSigning
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if err := (&info).Create(claOrgID, true); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "sign successfully"
}

// @Title Check
// @Description check whether contributor has signed cla
// @Param	platform	path 	string	true		"code platform"
// @Param	org		path 	string	true		"org"
// @Param	repo		path 	string	true		"repo"
// @Param	email		query 	string	true		"email"
// @Success 200
// @router /:platform/:org/:repo [get]
func (this *IndividualSigningController) Check() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	params := []string{":platform", ":org", ":repo", "email"}
	if err := checkAPIStringParameter(&this.Controller, params); err != nil {
		reason = err
		statusCode = 400
		return
	}

	v, err := models.IsIndividualSigned(
		this.GetString(":platform"),
		this.GetString(":org"),
		this.GetString(":repo"),
		this.GetString("email"),
	)
	if err != nil {
		reason = fmt.Errorf("Failed to check signing: %s", err.Error())
		statusCode = 500
		return
	}

	body = map[string]bool{
		"signed": v,
	}
}
