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
	doWhat := "add employee managers"

	pl, err := this.tokenPayloadOfCorpManager()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	opt, fr := this.fetchEmployeeManagerCreateOption(pl.Email)
	if fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}

	added, merr := opt.Create(pl.LinkID)
	if merr != nil {
		if merr.IsErrorOf(models.ErrNoLinkOrDuplicateManager) {
			this.sendFailedResponse(400, errDuplicateManager, merr, doWhat)
		} else {
			this.sendModelErrorAsResp(merr, doWhat)
		}
		return
	}

	this.sendResponse(doWhat+" successfully", 0)

	notifyCorpManagerWhenAdding(&pl.OrgInfo, added)
}

// @Title Delete
// @Description delete employee manager
// @Param	body		body 	models.EmployeeManagerCreateOption	true		"body for employee manager"
// @Success 204 {string} delete success!
// @router / [delete]
func (this *EmployeeManagerController) Delete() {
	doWhat := "delete employee managers"

	pl, err := this.tokenPayloadOfCorpManager()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	opt, fr := this.fetchEmployeeManagerCreateOption(pl.Email)
	if fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}

	deleted, merr := opt.Delete(pl.LinkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, doWhat)
		return
	}

	this.sendResponse(doWhat+"successfully", 0)

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
	doWhat := "list employee managers"

	pl, err := this.tokenPayloadOfCorpManager()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	r, merr := models.ListCorporationManagers(pl.LinkID, pl.Email, dbmodels.RoleManager)
	if merr != nil {
		this.sendModelErrorAsResp(merr, doWhat)
		return
	}

	this.sendResponse(r, 0)
}

func (this *EmployeeManagerController) fetchEmployeeManagerCreateOption(corpEmail string) (*models.EmployeeManagerCreateOption, *failedResult) {

	info := &models.EmployeeManagerCreateOption{}
	if err := this.fetchInputPayload(info); err != nil {
		return nil, newFailedResult(400, util.ErrInvalidParameter, err)
	}

	if merr := info.Validate(corpEmail); merr != nil {
		return nil, parseModelError(merr)
	}

	return info, nil
}
