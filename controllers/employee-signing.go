package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type EmployeeSigningController struct {
	beego.Controller
}

func (this *EmployeeSigningController) Prepare() {
	if getRequestMethod(&this.Controller) == http.MethodPost {
		// sign as employee
		apiPrepare(&this.Controller, []string{PermissionIndividualSigner})
	} else {
		// get, update and delete employee
		apiPrepare(&this.Controller, []string{PermissionEmployeeManager})
	}
}

// @Title Post
// @Description sign as employee
// @Param	:org_cla_id	path 	string				true		"org cla id"
// @Param	body		body 	models.IndividualSigning	true		"body for employee signing"
// @Success 201 {int} map
// @Failure util.ErrHasSigned		"employee has signed"
// @Failure util.ErrHasNotSigned	"corp has not signed"
// @Failure util.ErrSigningUncompleted	"corp has not been enabled"
// @router /:org_cla_id [post]
func (this *EmployeeSigningController) Post() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "sign as employee")
	}()

	orgCLAID, err := fetchStringParameter(&this.Controller, ":org_cla_id")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	ac, ec, err := getACOfCodePlatform(&this.Controller)
	if err != nil {
		reason = err
		errCode = ec
		statusCode = 400
		return
	}

	var info models.EmployeeSigning
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}
	if ec, err := (&info).Validate(orgCLAID, ac.Email); err != nil {
		reason = err
		errCode = ec
		statusCode = 400
		return
	}

	orgCLA := &models.OrgCLA{ID: orgCLAID}
	if err := orgCLA.Get(); err != nil {
		reason = err
		return
	}
	if isNotIndividualCLA(orgCLA) {
		reason = fmt.Errorf("invalid cla")
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	_, corpSign, err := models.GetCorporationSigningDetail(
		orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID, info.Email)
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

	orgRepo := buildOrgRepo(orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID)
	managers, err := models.ListCorporationManagers(&orgRepo, info.Email, dbmodels.RoleManager)
	if err != nil {
		reason = err
		return
	}
	if len(managers) == 0 {
		reason = fmt.Errorf("no managers")
		errCode = util.ErrNoCorpManager
		statusCode = 400
		return
	}

	cla := &models.CLA{ID: orgCLA.CLAID}
	if err := cla.GetFields(); err != nil {
		reason = err
		return
	}

	info.Info = getSingingInfo(info.Info, cla.Fields)

	err = (&info).Create(&orgRepo, false)
	if err != nil {
		reason = err
		return
	}
	body = "sign successfully"

	this.notifyManagers(managers, &info, orgCLA)
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

	var ac *acForCorpManagerPayload
	ac, errCode, reason = getACOfCorpManager(&this.Controller)
	if reason != nil {
		statusCode = 401
		return
	}

	opt := models.EmployeeSigningListOption{
		CLALanguage: this.GetString("cla_language"),
	}

	orgRepo := buildOrgRepo(ac.Platform, ac.OrgID, ac.RepoID)
	r, err := opt.List(&orgRepo, ac.Email)
	if err != nil {
		reason = err
		return
	}

	body = r
}

// @Title Update
// @Description enable/unable employee signing
// @Param	:email		path 	string	true		"email"
// @Success 202 {int} map
// @router /:email [put]
func (this *EmployeeSigningController) Update() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "enable/unable employee signing")
	}()

	employeeEmail, err := fetchStringParameter(&this.Controller, ":email")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	var ac *acForCorpManagerPayload
	ac, errCode, reason = getACOfCorpManager(&this.Controller)
	if reason != nil {
		statusCode = 401
		return
	}

	if !isSameCorp(ac.Email, employeeEmail) {
		reason = fmt.Errorf("not same corp")
		errCode = util.ErrNotSameCorp
		statusCode = 400
		return
	}

	var info models.EmployeeSigningUdateInfo
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	orgRepo := buildOrgRepo(ac.Platform, ac.OrgID, ac.RepoID)
	err = (&info).Update(&orgRepo, employeeEmail)
	if err != nil {
		reason = err
		return
	}

	body = "enabled employee successfully"

	msg := email.EmployeeNotification{
		Name:       employeeEmail,
		Manager:    ac.Email,
		ProjectURL: ac.orgRepoURL(),
		Org:        ac.OrgAlias,
	}
	subject := ""
	if info.Enabled {
		msg.Active = true
		subject = "Activate the CLA signing"
	} else {
		msg.Inactive = true
		subject = "Inavtivate the CLA signing"
	}
	sendEmailToIndividual(employeeEmail, ac.OrgEmail, subject, msg)
}

// @Title Delete
// @Description delete employee signing
// @Param	:email		path 	string	true		"email"
// @Success 204 {string} delete success!
// @router /:email [delete]
func (this *EmployeeSigningController) Delete() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body string

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "delete employee signing")
	}()

	employeeEmail, err := fetchStringParameter(&this.Controller, ":email")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	var ac *acForCorpManagerPayload
	ac, errCode, reason = getACOfCorpManager(&this.Controller)
	if reason != nil {
		statusCode = 401
		return
	}

	if !isSameCorp(ac.Email, employeeEmail) {
		reason = fmt.Errorf("not same corp")
		errCode = util.ErrNotSameCorp
		statusCode = 400
		return
	}

	orgRepo := buildOrgRepo(ac.Platform, ac.OrgID, ac.RepoID)
	err = models.DeleteEmployeeSigning(&orgRepo, employeeEmail)
	if err != nil {
		reason = err
		return
	}

	body = "delete employee successfully"

	msg := email.EmployeeNotification{
		Removing:   true,
		Name:       employeeEmail,
		Manager:    ac.Email,
		ProjectURL: ac.orgRepoURL(),
		Org:        ac.OrgAlias,
	}
	sendEmailToIndividual(employeeEmail, ac.OrgEmail, "Remove employee", msg)
}

func (this *EmployeeSigningController) notifyManagers(managers []dbmodels.CorporationManagerListResult, info *models.EmployeeSigning, orgCLA *models.OrgCLA) {
	ms := make([]string, 0, len(managers))
	to := make([]string, 0, len(managers))
	for _, item := range managers {
		to = append(to, item.Email)
		ms = append(ms, fmt.Sprintf("%s: %s", item.Name, item.Email))
	}

	msg := email.EmployeeSigning{
		Name:       info.Name,
		Org:        orgCLA.OrgAlias,
		ProjectURL: projectURL(orgCLA),
		Managers:   "  " + strings.Join(ms, "\n  "),
	}
	sendEmailToIndividual(
		info.Email, orgCLA.OrgEmail,
		fmt.Sprintf("Signing CLA on project of \"%s\"", msg.Org),
		msg,
	)

	msg1 := email.NotifyingManager{
		Org:              orgCLA.OrgAlias,
		EmployeeEmail:    info.Email,
		ProjectURL:       projectURL(orgCLA),
		URLOfCLAPlatform: conf.AppConfig.CLAPlatformURL,
	}
	sendEmail(to, orgCLA.OrgEmail, "An employee has signed CLA", msg1)
}
