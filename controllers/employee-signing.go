package controllers

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type EmployeeSigningController struct {
	baseController
}

func (this *EmployeeSigningController) Prepare() {
	if this.isPostRequest() {
		// sign as employee
		this.apiPrepare(PermissionIndividualSigner)
	} else {
		// get, update and delete employee
		this.apiPrepare(PermissionEmployeeManager)
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
	action := "sign employeee cla"
	sendResp := this.newFuncForSendingFailedResp(action)
	orgCLAID := this.GetString(":org_cla_id")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		sendResp(fr)
		return
	}

	var info models.EmployeeSigning
	if fr := this.fetchInputPayload(&info); fr != nil {
		sendResp(fr)
		return
	}
	if err := (&info).Validate(orgCLAID, pl.Email); err != nil {
		sendResp(parseModelError(err))
		return
	}

	orgCLA := &models.OrgCLA{ID: orgCLAID}
	if err := orgCLA.Get(); err != nil {
		sendResp(convertDBError1(err))
		return
	}
	if isNotIndividualCLA(orgCLA) {
		this.sendFailedResponse(400, util.ErrInvalidParameter, fmt.Errorf("invalid cla"), action)
		return
	}

	corpSignedCla, corpSign, err := models.GetCorporationSigningDetail(
		orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID, info.Email)
	if err != nil {
		sendResp(convertDBError1(err))
		return
	}

	if !corpSign.AdminAdded {
		this.sendFailedResponse(
			400, util.ErrSigningUncompleted, fmt.Errorf("the corp has not been enabled"), action,
		)
		return
	}

	managers, err := models.ListCorporationManagers(corpSignedCla, info.Email, dbmodels.RoleManager)
	if err != nil {
		sendResp(convertDBError1(err))
		return
	}
	if len(managers) == 0 {
		this.sendFailedResponse(400, util.ErrNoCorpManager, fmt.Errorf("no managers"), action)
		return
	}

	cla := &models.CLA{ID: orgCLA.CLAID}
	if err := cla.GetFields(); err != nil {
		sendResp(convertDBError1(err))
		return
	}

	info.Info = getSingingInfo(info.Info, cla.Fields)

	if err := (&info).Create(orgCLAID, false); err != nil {
		if err.IsErrorOf(models.ErrNoLinkOrResigned) {
			this.sendFailedResponse(400, errResigned, err, action)
		} else {
			sendResp(parseModelError(err))
		}
		return
	}

	this.sendSuccessResp("sign successfully")

	this.notifyManagers(managers, &info, orgCLA)
}

// @Title GetAll
// @Description get all the employees
// @Success 200 {int} map
// @router / [get]
func (this *EmployeeSigningController) GetAll() {
	sendResp := this.newFuncForSendingFailedResp("list employeee")

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	orgCLA := &models.OrgCLA{ID: pl.OrgCLAID}
	if err := orgCLA.Get(); err != nil {
		sendResp(convertDBError1(err))
		return
	}

	opt := models.OrgCLAListOption{
		Platform: orgCLA.Platform,
		OrgID:    orgCLA.OrgID,
		RepoID:   orgCLA.RepoID,
		ApplyTo:  dbmodels.ApplyToIndividual,
	}
	signings, err := opt.List()
	if err != nil {
		sendResp(convertDBError1(err))
		return
	}
	if len(signings) == 0 {
		return
	}

	r, merr := models.ListIndividualSigning(signings[0].ID, pl.Email, this.GetString("cla_language"))
	if merr != nil {
		sendResp(parseModelError(merr))
		return
	}

	this.sendSuccessResp(r)
}

// @Title Update
// @Description enable/unable employee signing
// @Param	:email		path 	string	true		"email"
// @Success 202 {int} map
// @router /:email [put]
func (this *EmployeeSigningController) Update() {
	action := "enable/unable employeee"
	sendResp := this.newFuncForSendingFailedResp(action)
	employeeEmail := this.GetString(":email")

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	if !pl.hasEmployee(employeeEmail) {
		this.sendFailedResponse(400, util.ErrNotSameCorp, fmt.Errorf("not same corp"), action)
		return
	}

	corpClaOrg := &models.OrgCLA{ID: pl.OrgCLAID}
	if err := corpClaOrg.Get(); err != nil {
		sendResp(convertDBError1(err))
		return
	}

	opt := models.OrgCLAListOption{
		Platform: corpClaOrg.Platform,
		OrgID:    corpClaOrg.OrgID,
		RepoID:   corpClaOrg.RepoID,
		ApplyTo:  dbmodels.ApplyToIndividual,
	}
	signings, err := opt.List()
	if err != nil {
		sendResp(convertDBError1(err))
		return
	}
	if len(signings) == 0 {
		return
	}

	var info models.EmployeeSigningUdateInfo
	if fr := this.fetchInputPayload(&info); fr != nil {
		sendResp(fr)
		return
	}

	if err := (&info).Update(signings[0].ID, employeeEmail); err != nil {
		if err.IsErrorOf(models.ErrNoLinkOrUnsigned) {
			this.sendFailedResponse(400, errUnsigned, err, action)
		} else {
			sendResp(parseModelError(err))
		}
		return
	}

	this.sendSuccessResp("enabled employee successfully")

	msg := email.EmployeeNotification{
		Name:       employeeEmail,
		Manager:    pl.Email,
		ProjectURL: projectURL(corpClaOrg),
		Org:        corpClaOrg.OrgAlias,
	}
	subject := ""
	if info.Enabled {
		msg.Active = true
		subject = "Activate the CLA signing"
	} else {
		msg.Inactive = true
		subject = "Inavtivate the CLA signing"
	}
	sendEmailToIndividual(employeeEmail, corpClaOrg.OrgEmail, subject, msg)
}

// @Title Delete
// @Description delete employee signing
// @Param	:email		path 	string	true		"email"
// @Success 204 {string} delete success!
// @router /:email [delete]
func (this *EmployeeSigningController) Delete() {
	action := "delete employee signing"
	sendResp := this.newFuncForSendingFailedResp(action)
	employeeEmail := this.GetString(":email")

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	if !pl.hasEmployee(employeeEmail) {
		this.sendFailedResponse(400, util.ErrNotSameCorp, fmt.Errorf("not same corp"), action)
		return
	}

	corpClaOrg := &models.OrgCLA{ID: pl.OrgCLAID}
	if err := corpClaOrg.Get(); err != nil {
		sendResp(convertDBError1(err))
		return
	}

	opt := models.OrgCLAListOption{
		Platform: corpClaOrg.Platform,
		OrgID:    corpClaOrg.OrgID,
		RepoID:   corpClaOrg.RepoID,
		ApplyTo:  dbmodels.ApplyToIndividual,
	}
	signings, err := opt.List()
	if err != nil {
		sendResp(convertDBError1(err))
		return
	}
	if len(signings) == 0 {
		return
	}

	if err := models.DeleteEmployeeSigning(signings[0].ID, employeeEmail); err != nil {
		sendResp(parseModelError(err))
		return
	}

	this.sendSuccessResp("delete employee successfully")

	msg := email.EmployeeNotification{
		Removing:   true,
		Name:       employeeEmail,
		Manager:    pl.Email,
		ProjectURL: projectURL(corpClaOrg),
		Org:        corpClaOrg.OrgAlias,
	}
	sendEmailToIndividual(employeeEmail, corpClaOrg.OrgEmail, "Remove employee", msg)
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
