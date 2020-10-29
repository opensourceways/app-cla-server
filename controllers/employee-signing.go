package controllers

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

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
		apiPrepare(&this.Controller, []string{PermissionIndividualSigner}, nil)
	} else {
		// get, update and delete employee
		apiPrepare(&this.Controller, []string{PermissionEmployeeManager}, nil)
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

	var info models.EmployeeSigning
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}
	if ec, err := (&info).Validate(orgCLAID); err != nil {
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

	corpSignedCla, corpSign, err := models.GetCorporationSigningDetail(
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

	cla := &models.CLA{ID: orgCLA.CLAID}
	if err := cla.GetFields(); err != nil {
		reason = err
		return
	}

	info.Info = getSingingInfo(info.Info, cla.Fields)

	err = (&info).Create(orgCLAID, orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID, false)
	if err != nil {
		reason = err
		return
	}
	body = "sign successfully"

	msg := email.EmployeeSigning{
		Name: info.Name,
		Org:  orgCLA.OrgID,
		Repo: orgCLA.RepoID,
	}
	sendEmailToIndividual(info.Email, orgCLA.OrgEmail, "Signing as employee", msg)

	this.notifyManagers(corpSignedCla, &info, orgCLA)
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

	orgCLAID, corpEmail, err := parseCorpManagerUser(&this.Controller)
	if err != nil {
		reason = err
		errCode = util.ErrUnknownToken
		statusCode = 401
		return
	}

	orgCLA := &models.OrgCLA{ID: orgCLAID}
	if err := orgCLA.Get(); err != nil {
		reason = err
		return
	}

	opt := models.EmployeeSigningListOption{
		CLALanguage: this.GetString("cla_language"),
	}

	r, err := opt.List(corpEmail, orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID)
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

	corpClaOrgID := ""
	statusCode, errCode, corpClaOrgID, reason = this.canHandleOnEmployee(employeeEmail)
	if reason != nil {
		return
	}

	corpClaOrg := &models.OrgCLA{ID: corpClaOrgID}
	if err := corpClaOrg.Get(); err != nil {
		reason = err
		return
	}

	var info models.EmployeeSigningUdateInfo
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	err = (&info).Update(corpClaOrg.Platform, corpClaOrg.OrgID, corpClaOrg.RepoID, employeeEmail)
	if err != nil {
		reason = err
		return
	}

	body = "enabled employee successfully"

	msg := email.EmployeeNotification{}
	subject := ""
	if info.Enabled {
		msg.Active = true
		subject = "Activate employee"
	} else {
		msg.Inactive = true
		subject = "Inavtivate employee"
	}
	sendEmailToIndividual(employeeEmail, corpClaOrg.OrgEmail, subject, msg)
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

	corpClaOrgID := ""
	statusCode, errCode, corpClaOrgID, reason = this.canHandleOnEmployee(employeeEmail)
	if reason != nil {
		return
	}

	corpClaOrg := &models.OrgCLA{ID: corpClaOrgID}
	if err := corpClaOrg.Get(); err != nil {
		reason = err
		return
	}

	err = models.DeleteEmployeeSigning(corpClaOrg.Platform, corpClaOrg.OrgID, corpClaOrg.RepoID, employeeEmail)
	if err != nil {
		reason = err
		return
	}

	body = "delete employee successfully"

	msg := email.EmployeeNotification{Removing: true}
	sendEmailToIndividual(employeeEmail, corpClaOrg.OrgEmail, "Remove employee", msg)
}

func (this *EmployeeSigningController) canHandleOnEmployee(employeeEmail string) (int, string, string, error) {
	corpClaOrgID, corpEmail, err := parseCorpManagerUser(&this.Controller)
	if err != nil {
		return 401, util.ErrUnknownToken, "", err
	}

	if !isSameCorp(corpEmail, employeeEmail) {
		return 400, util.ErrNotSameCorp, "", fmt.Errorf("not same corp")
	}

	return 0, "", corpClaOrgID, nil
}

func (this *EmployeeSigningController) notifyManagers(corpClaOrgID string, info *models.EmployeeSigning, orgCLA *models.OrgCLA) {
	managers, err := models.ListCorporationManagers(corpClaOrgID, info.Email, dbmodels.RoleManager)
	if err != nil {
		beego.Error(err)
		return
	}

	if len(managers) == 0 {
		return
	}

	to := make([]string, 0, len(managers))
	for _, item := range managers {
		if item.Role == dbmodels.RoleManager {
			to = append(to, item.Email)
		}
	}

	msg := email.NotifyingManager{
		Name:     info.Name,
		Platform: orgCLA.Platform,
		Org:      orgCLA.OrgID,
		Repo:     orgCLA.RepoID,
	}
	sendEmail(to, orgCLA.OrgEmail, "A employee has signed CLA", msg)
}
