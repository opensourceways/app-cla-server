package controllers

import (
	"fmt"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/models"
)

type CLAController struct {
	beego.Controller
}

func (this *CLAController) Prepare() {
	apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, &acForCodePlatformPayload{})
}

// @Title Delete CLA
// @Description delete cla
// @Param	uid		path 	string	true		"cla id"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:uid [delete]
func (this *CLAController) Delete() {
	var statusCode = 0
	var reason error
	var body string

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	uid := this.GetString(":uid")
	if uid == "" {
		reason = fmt.Errorf("missing cla id")
		statusCode = 400
		return
	}

	cla := models.CLA{ID: uid}

	if err := (&cla).Delete(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "delete cla successfully"
}

// @Title Get
// @Description get cla by uid
// @Param	uid		path 	string	true		"The key for cla"
// @Success 200 {object} models.CLA
// @Failure 403 :uid is empty
// @router /:uid [get]
func (this *CLAController) Get() {
	var statusCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	uid := this.GetString(":uid")
	if uid == "" {
		reason = fmt.Errorf("missing cla id")
		statusCode = 400
		return
	}

	cla := models.CLA{ID: uid}

	if err := (&cla).Get(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = cla
}

// @Title GetAllCLA
// @Description get all clas
// @Success 200 {object} models.CLA
// @router / [get]
func (this *CLAController) GetAll() {
	var statusCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	/*
		user, err := getApiAccessUser(&this.Controller)
		if err != nil {
			reason = err
			statusCode = 400
			return
		}
	*/

	clas := models.CLAListOptions{
		// Submitter: user,
		Name:     this.GetString("name"),
		ApplyTo:  this.GetString("apply_to"),
		Language: this.GetString("language"),
	}

	r, err := clas.Get()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = r
}
