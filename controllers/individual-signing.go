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
		reason = fmt.Errorf("missing cla_org_id")
		statusCode = 400
		return
	}

	var info models.IndividualSigning
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if err := (&info).Create(claOrgID); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "sign successfully"
}
