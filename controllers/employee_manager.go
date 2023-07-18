package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
)

type EmployeeManagerController struct {
	baseController
}

func (ctl *EmployeeManagerController) Prepare() {
	ctl.apiPrepare(PermissionCorpAdmin)
}

// @Title Post
// @Description add employee managers
// @Tags EmployeeManager
// @Accept json
// @Param  body  body  models.EmployeeManagerCreateOption  true  "body for employee manager"
// @Success 201 {object} controllers.respData
// @router / [post]
func (ctl *EmployeeManagerController) Post() {
	action := "corp admin adds employee managers"
	sendResp := ctl.newFuncForSendingFailedResp(action)

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	orgInfo, merr := models.GetLink(pl.LinkID)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)

		return
	}

	info := &models.EmployeeManagerCreateOption{}
	if fr := ctl.fetchInputPayload(info); fr != nil {
		sendResp(fr)
		return
	}

	added, merr := models.AddEmployeeManager(pl.SigningId, info)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
		return
	}

	ctl.sendSuccessResp(action, "successfully")

	notifyCorpManagerWhenAdding(&orgInfo, added)
}

// @Title Delete
// @Description delete employee manager
// @Tags EmployeeManager
// @Accept json
// @Param  body  body  models.EmployeeManagerDeleteOption  true  "body for employee manager"
// @Success 204 {object} controllers.respData
// @router / [delete]
func (ctl *EmployeeManagerController) Delete() {
	action := "corp admin deletes employee managers"
	sendResp := ctl.newFuncForSendingFailedResp(action)

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	orgInfo, merr := models.GetLink(pl.LinkID)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)

		return
	}

	info := &models.EmployeeManagerDeleteOption{}
	if fr := ctl.fetchInputPayload(info); fr != nil {
		sendResp(fr)
		return
	}

	deleted, merr := models.RemoveEmployeeManager(pl.SigningId, info)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)
		return
	}

	ctl.sendSuccessResp(action, "successfully")

	subject := fmt.Sprintf("Revoking the authorization on project of \"%s\"", orgInfo.OrgAlias)

	for _, item := range deleted {
		msg := emailtmpl.RemovingCorpManager{
			User:       item.Name,
			Org:        orgInfo.OrgAlias,
			ProjectURL: orgInfo.ProjectURL(),
		}
		sendEmailToIndividual(item.Email, &orgInfo, subject, &msg)
	}
}

// @Title GetAll
// @Description get all employee managers
// @Tags EmployeeManager
// @Accept json
// @Success 200 {object} models.CorporationManagerListResult
// @router / [get]
func (ctl *EmployeeManagerController) GetAll() {
	action := "corp admin lists employee managers"
	sendResp := ctl.newFuncForSendingFailedResp(action)

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	r, err := models.ListEmployeeManagers(pl.SigningId)
	if err == nil {
		ctl.sendSuccessResp(action, r)
	} else {
		sendResp(parseModelError(err))
	}
}
