package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
	"github.com/opensourceways/app-cla-server/worker"
)

type EmployeeSigningController struct {
	beego.Controller
}

func (this *EmployeeSigningController) Prepare() {
	if this.Ctx.Request.Method == http.MethodPost {
		// sign as employee
		apiPrepare(&this.Controller, []string{PermissionIndividualSigner}, nil)
	} else {
		// get, update and delete employee
		apiPrepare(&this.Controller, []string{PermissionEmployeeManager}, nil)
	}
}

// @Title Post
// @Description sign as employee
// @Param	:cla_org_id	path 	string				true		"cla org id"
// @Param	body		body 	models.IndividualSigning	true		"body for employee signing"
// @Success 201 {int} map
// @Failure util.ErrHasSigned		"employee has signed"
// @Failure util.ErrHasNotSigned	"corp has not signed"
// @Failure util.ErrSigningUncompleted	"corp has not been enabled"
// @router /:cla_org_id [post]
func (this *EmployeeSigningController) Post() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "sign as employee")
	}()

	claOrgID, err := fetchStringParameter(&this.Controller, ":cla_org_id")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	var info models.IndividualSigning
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	claOrg := &models.CLAOrg{ID: claOrgID}
	if err := claOrg.Get(); err != nil {
		reason = err
		return
	}

	emailCfg := &models.OrgEmail{Email: claOrg.OrgEmail}
	if err := emailCfg.Get(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	corpSignedCla, corpSign, err := models.GetCorporationSigningDetail(
		claOrg.Platform, claOrg.OrgID, claOrg.RepoID, info.Email)
	if err != nil {
		reason = err
		return
	}

	if !corpSign.AdminAdded {
		reason = fmt.Errorf("the corp has not been enabled")
		errCode = util.ErrSigningUncompleted
		statusCode = 400
		return
	}

	managers, err := models.ListCorporationManagers(corpSignedCla, info.Email, dbmodels.RoleManager)
	if err != nil {
		reason = err
		return
	}

	err = (&info).Create(claOrgID, claOrg.Platform, claOrg.OrgID, claOrg.RepoID, false)
	if err != nil {
		reason = err
		return
	}
	body = "sign successfully"

	if len(managers) > 0 {
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
		worker.GetEmailWorker().SendSimpleMessage(emailCfg, &msg)
	}
}

// @Title GetAll
// @Description get all the employees
// @Param	:platform	path 	string	true		"code platform"
// @Param	:org		path 	string	true		"org"
// @Success 200 {int} map
// @router /:platform/:org [get]
func (this *EmployeeSigningController) GetAll() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "list employees")
	}()

	params := []string{":platform", ":org"}
	if err := checkAPIStringParameter(&this.Controller, params); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	opt := models.EmployeeSigningListOption{
		RepoID:      this.GetString("repo_id"),
		CLALanguage: this.GetString("cla_language"),
	}

	email, err := getApiAccessUser(&this.Controller)
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	r, err := opt.List(email, this.GetString(":platform"), this.GetString(":org"))
	if err != nil {
		reason = err
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
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "enable/unable employee signing")
	}()

	if err := checkAPIStringParameter(&this.Controller, []string{":cla_org_id", ":email"}); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return

	}
	employeeEmail := this.GetString(":email")

	statusCode, errCode, reason = checkSameCorp(&this.Controller, employeeEmail)
	if reason != nil {
		return
	}

	var info models.EmployeeSigningUdateInfo
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	err := (&info).Update(this.GetString(":cla_org_id"), employeeEmail)
	if err != nil {
		reason = err
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
	var statusCode = 0
	var errCode = ""
	var reason error
	var body string

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "delete employee signing")
	}()

	if err := checkAPIStringParameter(&this.Controller, []string{":cla_org_id", ":email"}); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return

	}
	employeeEmail := this.GetString(":email")

	statusCode, errCode, reason = checkSameCorp(&this.Controller, employeeEmail)
	if reason != nil {
		return
	}

	err := models.DeleteEmployeeSigning(this.GetString(":cla_org_id"), this.GetString(":email"))
	if err != nil {
		reason = err
		statusCode, errCode = convertDBError(err)
		return
	}

	body = "delete employee successfully"
}
