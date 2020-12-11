package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type EmployeeManagerController struct {
	baseController
}

func (this *EmployeeManagerController) Prepare() {
	this.apiPrepare(PermissionCorporAdmin)
}

// @Title Post
// @Description add employee managers
// @Param	body		body 	models.EmployeeManagerCreateOption	true		"body for employee manager"
// @Success 201 {int} map
// @router / [post]
func (this *EmployeeManagerController) Post() {
	this.addOrDeleteManagers(true)
}

// @Title Delete
// @Description delete employee manager
// @Param	body		body 	models.EmployeeManagerCreateOption	true		"body for employee manager"
// @Success 204 {string} delete success!
// @router / [delete]
func (this *EmployeeManagerController) Delete() {
	this.addOrDeleteManagers(false)
}

// @Title GetAll
// @Description get all employee managers
// @Success 200 {object} dbmodels.CorporationManagerListResult
// @router / [get]
func (this *EmployeeManagerController) GetAll() {
	doWhat := "list employee managers"

	pl, err := this.tokenPayloadOfCorpManager()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	r, err := models.ListCorporationManagers(pl.OrgCLAID, pl.Email, dbmodels.RoleManager)
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	this.sendResponse(r, 0)
}

func (this *EmployeeManagerController) addOrDeleteManagers(toAdd bool) {
	doWhat := fmt.Sprintf("add/remove employee managers")
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		op := "add"
		if !toAdd {
			op = "delete"
		}
		body = fmt.Sprintf("%s employee manager successfully", op)

		sendResponse(&this.Controller, statusCode, errCode, reason, body, fmt.Sprintf("%s employee managers", op))
	}()

	pl, err := this.tokenPayloadOfCorpManager()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	var info models.EmployeeManagerCreateOption
	if err := this.fetchInputPayload(&info); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, doWhat)
		return
	}

	if c, err := (&info).Validate(pl.Email); err != nil {
		this.sendFailedResponse(400, c, err, doWhat)
		return
	}

	orgCLA := &models.OrgCLA{ID: pl.OrgCLAID}
	if err := orgCLA.Get(); err != nil {
		reason = err
		return
	}

	if toAdd {
		added, err := (&info).Create(pl.OrgCLAID)
		if err != nil {
			this.sendFailedResponse(0, "", err, doWhat)
			return
		}
		notifyCorpManagerWhenAdding(orgCLA, added)
	} else {
		deleted, err := (&info).Delete(pl.OrgCLAID)
		if err != nil {
			this.sendFailedResponse(0, "", err, doWhat)
			return
		}

		subject := fmt.Sprintf("Revoking the authorization on project of \"%s\"", orgCLA.OrgAlias)
		for _, item := range deleted {
			msg := email.RemovingCorpManager{
				User:       item.Name,
				Org:        orgCLA.OrgAlias,
				ProjectURL: projectURL(orgCLA),
			}
			sendEmailToIndividual(item.Email, orgCLA.OrgEmail, subject, msg)
		}
	}
}
