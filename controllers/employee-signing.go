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
	if getRequestMethod(&this.Controller) == http.MethodPost {
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
// @Success 200 {int} map
// @router / [get]
func (this *EmployeeSigningController) GetAll() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "list employees")
	}()

	claOrgID, corpEmail, err := parseCorpManagerUser(&this.Controller)
	if err != nil {
		reason = err
		errCode = util.ErrUnknownToken
		statusCode = 401
		return
	}

	claOrg := &models.CLAOrg{ID: claOrgID}
	if err := claOrg.Get(); err != nil {
		reason = err
		return
	}

	opt := models.EmployeeSigningListOption{
		CLALanguage: this.GetString("cla_language"),
	}

	r, err := opt.List(corpEmail, claOrg.Platform, claOrg.OrgID, claOrg.RepoID)
	if err != nil {
		reason = err
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
	claOrgID := this.GetString(":cla_org_id")

	statusCode, errCode, reason = this.canHandleOnEmployee(claOrgID, employeeEmail)
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

	if err := (&info).Update(claOrgID, employeeEmail); err != nil {
		reason = err
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
	claOrgID := this.GetString(":cla_org_id")

	statusCode, errCode, reason = this.canHandleOnEmployee(claOrgID, employeeEmail)
	if reason != nil {
		return
	}

	if err := models.DeleteEmployeeSigning(claOrgID, employeeEmail); err != nil {
		reason = err
		return
	}

	body = "delete employee successfully"
}

func (this *EmployeeSigningController) canHandleOnEmployee(claOrgID, employeeEmail string) (int, string, error) {
	corpClaOrgID, corpEmail, err := parseCorpManagerUser(&this.Controller)
	if err != nil {
		return 401, util.ErrUnknownToken, err
	}

	if !isSameCorp(corpEmail, employeeEmail) {
		return 400, util.ErrNotSameCorp, fmt.Errorf("not same corp")
	}

	claOrg := &models.CLAOrg{ID: claOrgID}
	if err := claOrg.Get(); err != nil {
		return 0, "", err
	}

	corpClaOrg := &models.CLAOrg{ID: corpClaOrgID}
	if err := corpClaOrg.Get(); err != nil {
		return 0, "", err
	}

	if claOrg.Platform != corpClaOrg.Platform ||
		claOrg.OrgID != corpClaOrg.OrgID ||
		claOrg.RepoID != corpClaOrg.RepoID {
		return 400, util.ErrInvalidParameter, fmt.Errorf("not the same repo")
	}

	return 0, "", nil
}
