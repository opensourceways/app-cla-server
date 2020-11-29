package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type CLAController struct {
	baseController
}

func (this *CLAController) Prepare() {
	this.apiPrepare(PermissionOwnerOfOrg)
}

// @Title Link
// @Description link org and cla
// @Param	body		body 	models.OrgCLA	true		"body for org-repo content"
// @Success 201 {int} models.OrgCLA
// @Failure 403 body is empty
// @router /:org_id/:apply_to [post]
func (this *CLAController) AddCLA() {
	doWhat := "add cla"

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	org, repo := parseOrgAndRepo(this.GetString(":org_id"))
	if r := isOwnerOfOrg(pl, org); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, doWhat)
		return
	}

	var input interface {
		AddCLA(orgRepo *dbmodels.OrgRepo) error
		Validate() (string, error)
	}
	if this.GetString(":apply_to") == dbmodels.ApplyToIndividual {
		input = &models.CLACreateOption{}
	} else {
		input = &models.CorpCLACreateOption{}
	}

	if err := this.fetchInputPayload(input); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, doWhat)
		return
	}

	if ec, err := input.Validate(); err != nil {
		this.sendFailedResponse(400, ec, err, doWhat)
		return
	}

	orgRepo := buildOrgRepo(pl.Platform, org, repo)
	if err := input.AddCLA(&orgRepo); err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	this.sendResponse("add cla successfully", 0)
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
