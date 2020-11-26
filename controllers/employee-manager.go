package controllers

import (
	"fmt"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type EmployeeManagerController struct {
	beego.Controller
}

func (this *EmployeeManagerController) Prepare() {
	apiPrepare(&this.Controller, []string{PermissionCorporAdmin})
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
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "list employee managers")
	}()

	var ac *acForCorpManagerPayload
	ac, errCode, reason = getACOfCorpManager(&this.Controller)
	if reason != nil {
		statusCode = 401
		return
	}

	orgRepo := buildOrgRepo(ac.Platform, ac.OrgID, ac.RepoID)
	r, err := models.ListCorporationManagers(&orgRepo, ac.Email, dbmodels.RoleManager)
	if err != nil {
		reason = err
		return
	}

	body = r
}

func (this *EmployeeManagerController) addOrDeleteManagers(toAdd bool) {
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

	var ac *acForCorpManagerPayload
	ac, errCode, reason = getACOfCorpManager(&this.Controller)
	if reason != nil {
		statusCode = 401
		return
	}

	var info models.EmployeeManagerCreateOption
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	if c, err := (&info).Validate(ac.Email); err != nil {
		reason = err
		errCode = c
		statusCode = 400
		return
	}

	orgRepo := buildOrgRepo(ac.Platform, ac.OrgID, ac.RepoID)
	if toAdd {
		added, err := (&info).Create(&orgRepo)
		if err != nil {
			reason = err
		} else {
			notifyCorpManagerWhenAdding(ac.OrgAlias, ac.orgRepoURL(), ac.OrgEmail, added)
		}

	} else {
		deleted, err := (&info).Delete(&orgRepo)
		if err != nil {
			reason = err
		} else {
			subject := fmt.Sprintf("Revoking the authorization on project of \"%s\"", ac.OrgAlias)

			for _, item := range deleted {
				msg := email.RemovingCorpManager{
					User:       item.Name,
					Org:        ac.OrgAlias,
					ProjectURL: ac.orgRepoURL(),
				}
				sendEmailToIndividual(item.Email, ac.OrgEmail, subject, msg)
			}
		}
	}
}
