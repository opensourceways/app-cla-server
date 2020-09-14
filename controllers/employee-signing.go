package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/worker"
)

type EmployeeSigningController struct {
	beego.Controller
}

func (this *EmployeeSigningController) Prepare() {
	if this.Ctx.Request.Method == http.MethodPost {
		apiPrepare(&this.Controller, []string{PermissionIndividualSigner})
	} else {
		apiPrepare(&this.Controller, []string{PermissionEmployeeManager})
	}
}

// @Title Employee signing
// @Description sign as employee
// @Param	body		body 	models.EmployeeSigning	true		"body for employee signing"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router / [post]
func (this *EmployeeSigningController) Post() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.EmployeeSigning
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}

	claOrg := &models.CLAOrg{ID: info.CLAOrgID}
	if err := claOrg.Get(); err != nil {
		reason = err
		statusCode = 400
		return
	}

	emailInfo := &models.OrgEmail{Email: claOrg.OrgEmail}
	if err := emailInfo.Get(); err != nil {
		reason = err
		statusCode = 400
		return
	}

	opt := models.CLAOrgListOption{
		Platform: claOrg.Platform,
		OrgID:    claOrg.OrgID,
		RepoID:   claOrg.RepoID,
		ApplyTo:  dbmodels.ApplyToCorporation,
	}
	claOrgs, err := opt.List()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}
	if len(claOrgs) == 0 {
		reason = fmt.Errorf("this org has not been bound any cla to be signed as corporation")
		statusCode = 400
		return
	}

	ids := make([]string, 0, len(claOrgs))
	for _, i := range claOrgs {
		ids = append(ids, i.ID)
	}
	managers, err := models.ListManagersWhenEmployeeSigning(ids, info.Email)
	if err != nil {
		reason = err
		statusCode = 500
		return
	}
	if managers == nil || len(managers) == 0 {
		reason = fmt.Errorf("the corporation has not signed")
		statusCode = 500
		return
	}

	if err := (&info).Create(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	msg := email.EmailMessage{
		To:      []string{},
		Subject: "Notification",
		Content: "somebody has signed",
	}
	for _, item := range managers {
		if item.Role == dbmodels.RoleManager {
			msg.To = append(msg.To, item.Email)
		}
	}
	if len(msg.To) > 0 {
		worker.GetEmailWorker().SendSimpleMessage(emailInfo, &msg)
	}
	body = "sign successfully"
}

// @Title GetAll
// @Description get all the employees
// @Success 200 {object} models.EmployeeSigning
// @router / [get]
func (this *EmployeeSigningController) GetAll() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	opt := models.EmployeeSigningListOption{
		Platform:         this.GetString("platform"),
		OrgID:            this.GetString("org_id"),
		RepoID:           this.GetString("repo_id"),
		CLALanguage:      this.GetString("cla_language"),
		CorporationEmail: this.GetString("corporation_email"),
	}

	r, err := opt.List()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = r
}

// @Title Enable employee signing
// @Description enable employee
// @Param	body		body 	models.EmployeeSigning	true		"body for employee signing"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router / [put]
func (this *EmployeeSigningController) Update() {
	var statusCode = 202
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.EmployeeSigningUdateInfo
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if err := (&info).Update(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "enabled employee successfully"
}
