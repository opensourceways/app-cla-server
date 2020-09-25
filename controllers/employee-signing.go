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
		apiPrepare(&this.Controller, []string{PermissionIndividualSigner}, nil)
	} else {
		apiPrepare(&this.Controller, []string{PermissionEmployeeManager}, nil)
	}
}

// @Title Employee signing
// @Description sign as employee
// @Param	body		body 	models.EmployeeSigning	true		"body for employee signing"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router /:cla_org_id [post]
func (this *EmployeeSigningController) Post() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	claOrgID := this.GetString(":cla_org_id")
	if claOrgID == "" {
		reason = fmt.Errorf("missing cla_org_id")
		statusCode = 400
		return
	}

	var info models.EmployeeSigning
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}
	info.CLAOrgID = claOrgID

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

	if err := (&info).Create(claOrgID); err != nil {
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
// @router / [get]
func (this *EmployeeSigningController) GetAll() {
	var statusCode = 200
	var errCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body)
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
		reason = fmt.Errorf("Failed to list employees, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	body = r
}

// @Title Update
// @Description enable/unable employee signing
// @Param	:cla_org_id	path 	string	true		"cla org id"
// @Param	:email		path 	string	true		"email"
// @Success 202 {int} map
// @router /:cla_org_id/:email [put]
func (this *EmployeeSigningController) Update() {
	var statusCode = 202
	var errCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body)
	}()

	if err := checkAPIStringParameter(&this.Controller, []string{":cla_org_id", ":email"}); err != nil {
		reason = err
		errCode = ErrInvalidParameter
		statusCode = 400
		return

	}

	var info models.EmployeeSigningUdateInfo
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = ErrInvalidParameter
		statusCode = 400
		return
	}

	err := (&info).Update(this.GetString(":cla_org_id"), this.GetString(":email"))
	if err != nil {
		reason = fmt.Errorf("Failed to update employee signing, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	body = "enabled employee successfully"
}

// @Title Delete
// @Description delete employee signing
// @Param	:cla_org_id	path 	string	true		"cla org id"
// @Param	:email		path 	string	true		"email"
// @Success 204 {string} delete success!
// @router /:cla_org_id/:email [delete]
func (this *EmployeeSigningController) Delete() {
	var statusCode = 204
	var errCode = 0
	var reason error
	var body string

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body)
	}()

	if err := checkAPIStringParameter(&this.Controller, []string{":cla_org_id", ":email"}); err != nil {
		reason = err
		errCode = ErrInvalidParameter
		statusCode = 400
		return

	}

	err := models.DeleteEmployeeSigning(this.GetString(":cla_org_id"), this.GetString(":email"))
	if err != nil {
		reason = fmt.Errorf("Failed to delete employee signing, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	body = "delete employee successfully"
}
