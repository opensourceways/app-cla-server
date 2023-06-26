package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
)

type EmployeeManagerController struct {
	baseController
}

func (this *EmployeeManagerController) Prepare() {
	this.stopRunIfSignSerivceIsUnabled()

	this.apiPrepare(PermissionCorpAdmin)
}

// @Title Post
// @Description add employee managers
// @Param	body		body 	models.EmployeeManagerCreateOption	true		"body for employee manager"
// @Success 201 {int} map
// @router / [post]
func (this *EmployeeManagerController) Post() {
	action := "add employee managers"
	sendResp := this.newFuncForSendingFailedResp(action)

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	info := &models.EmployeeManagerCreateOption{}
	if fr := this.fetchInputPayload(info); fr != nil {
		sendResp(fr)
		return
	}

	// TODO csid
	added, merr := info.Add("")
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendSuccessResp(action + " successfully")

	notifyCorpManagerWhenAdding(&pl.OrgInfo, added)
}

// @Title Delete
// @Description delete employee manager
// @Param	body		body 	models.EmployeeManagerCreateOption	true		"body for employee manager"
// @Success 204 {string} delete success!
// @router / [delete]
func (this *EmployeeManagerController) Delete() {
	action := "delete employee managers"
	sendResp := this.newFuncForSendingFailedResp(action)

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	info := &models.EmployeeManagerCreateOption{}
	if fr := this.fetchInputPayload(info); fr != nil {
		sendResp(fr)
		return
	}

	domains, fr := listCorpEmailDomain(pl.LinkID, pl.Email)
	if fr != nil {
		fr.statusCode = 500
		sendResp(fr)
		return
	}

	if err := info.ValidateWhenDeleting(pl.Email, domains); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	deleted, merr := info.Delete(pl.LinkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendSuccessResp(action + "successfully")

	subject := fmt.Sprintf("Revoking the authorization on project of \"%s\"", pl.OrgAlias)

	for _, item := range deleted {
		msg := email.RemovingCorpManager{
			User:       item.Name,
			Org:        pl.OrgAlias,
			ProjectURL: pl.ProjectURL(),
		}
		sendEmailToIndividual(item.Email, pl.OrgEmail, subject, msg)
	}
}

// @Title GetAll
// @Description get all employee managers
// @Success 200 {object} dbmodels.CorporationManagerListResult
// @router / [get]
func (this *EmployeeManagerController) GetAll() {
	sendResp := this.newFuncForSendingFailedResp("list employee managers")

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	r, err := models.ListCorporationManagers(pl.LinkID, pl.Email, dbmodels.RoleManager)
	if err == nil {
		this.sendSuccessResp(r)
	} else {
		sendResp(parseModelError(err))
	}
}
