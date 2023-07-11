package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
)

type EmployeeManagerController struct {
	baseController
}

func (this *EmployeeManagerController) Prepare() {
	this.apiPrepare(PermissionCorpAdmin)
}

// @Title Post
// @Description add employee managers
// @Tags EmployeeManager
// @Accept json
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

	orgInfo, merr := models.GetLink(pl.LinkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)

		return
	}

	info := &models.EmployeeManagerCreateOption{}
	if fr := this.fetchInputPayload(info); fr != nil {
		sendResp(fr)
		return
	}

	added, merr := models.AddEmployeeManager(pl.SigningId, info)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendSuccessResp(action + " successfully")

	notifyCorpManagerWhenAdding(&orgInfo, added)
}

// @Title Delete
// @Description delete employee manager
// @Tags EmployeeManager
// @Accept json
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

	orgInfo, merr := models.GetLink(pl.LinkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)

		return
	}

	info := &models.EmployeeManagerCreateOption{}
	if fr := this.fetchInputPayload(info); fr != nil {
		sendResp(fr)
		return
	}

	deleted, merr := models.RemoveEmployeeManager(pl.SigningId, info)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendSuccessResp(action + "successfully")

	subject := fmt.Sprintf("Revoking the authorization on project of \"%s\"", orgInfo.OrgAlias)

	for _, item := range deleted {
		msg := emailtmpl.RemovingCorpManager{
			User:       item.Name,
			Org:        orgInfo.OrgAlias,
			ProjectURL: orgInfo.ProjectURL(),
		}
		sendEmailToIndividual(item.Email, &orgInfo, subject, msg)
	}
}

// @Title GetAll
// @Description get all employee managers
// @Tags EmployeeManager
// @Accept json
// @Success 200 {object} dbmodels.CorporationManagerListResult
// @router / [get]
func (this *EmployeeManagerController) GetAll() {
	sendResp := this.newFuncForSendingFailedResp("list employee managers")

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	r, err := models.ListEmployeeManagers(pl.SigningId)
	if err == nil {
		this.sendSuccessResp(r)
	} else {
		sendResp(parseModelError(err))
	}
}
