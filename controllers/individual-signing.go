package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/models"
)

type IndividualSigningController struct {
	beego.Controller
}

func (this *IndividualSigningController) Prepare() {
	apiPrepare(&this.Controller, []string{PermissionIndividualSigner})
}

// @Title Individual signing
// @Description sign as individual
// @Param	body		body 	models.IndividualSigning	true		"body for individual signing"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router / [post]
func (this *IndividualSigningController) Post() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.IndividualSigning
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
