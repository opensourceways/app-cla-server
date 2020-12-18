package controllers

import (
	"fmt"
	"net/http"
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
	if getRequestMethod(&this.Controller) == http.MethodPost {
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
// @router /:link_id/:cla_lang/:cla_hash [post]
func (this *EmployeeSigningController) Post() {
	action := "sign as employee"
	linkID := this.GetString(":link_id")
	claLang := this.GetString(":cla_lang")

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, action)
		return
	}

	var info models.EmployeeSigning
	if err := this.fetchInputPayload(&info); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, action)
		return
	}
	info.CLALanguage = claLang

	if merr := (&info).Validate(linkID, pl.Email); err != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	orgInfo, merr := models.GetOrgOfLink(linkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	// TODO: return orgInfo by ListCorporationManagers
	managers, merr := models.ListCorporationManagers(linkID, info.Email, "")
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}
	if len(managers) <= 1 {
		this.sendFailedResponse(400, ErrNoCorpEmployeeManager, fmt.Errorf("no managers"), action)
		return
	}

	claInfo, merr := models.GetCLAInfoSigned(linkID, claLang, dbmodels.ApplyToIndividual)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}
	if claInfo == nil {
		// no contributor signed for this language. lock to avoid the cla to be changed
		// before writing to the db.
		unlock, err := util.Lock(genOrgFileLockPath(orgInfo.Platform, orgInfo.OrgID, orgInfo.RepoID))
		if err != nil {
			this.sendFailedResponse(500, util.ErrSystemError, err, action)
			return
		}
		defer unlock()

		claInfo, merr = models.GetCLAInfoToSign(linkID, claLang, dbmodels.ApplyToIndividual)
		if merr != nil {
			this.sendModelErrorAsResp(merr, action)
			return
		}
	}

	if claInfo.CLAHash != this.GetString(":cla_hash") {
		this.sendFailedResponse(400, errUnmatchedCLA, fmt.Errorf("invalid cla"), action)
		return
	}

	info.Info = getSingingInfo(info.Info, claInfo.Fields)

	if merr := (&info).Create(linkID, false); merr != nil {
		if merr.IsErrorOf(models.ErrNoLinkOrResign) {
			this.sendFailedResponse(400, errHasSigned, merr, action)
		} else {
			this.sendModelErrorAsResp(merr, action)
		}
		return
	}

	this.sendResponse("sign successfully", 0)

	this.notifyManagers(managers, &info, orgInfo)
}

// @Title GetAll
// @Description get all the employees
// @Success 200 {int} map
// @router / [get]
func (this *EmployeeSigningController) GetAll() {
	action := "list employees"

	pl, err := this.tokenPayloadOfCorpManager()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, action)
		return
	}

	opt := models.EmployeeSigningListOption{
		CLALanguage: this.GetString("cla_language"),
	}

	r, merr := opt.List(pl.LinkID, pl.Email)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendResponse(r, 0)
}

// @Title Update
// @Description enable/unable employee signing
// @Param	:email		path 	string	true		"email"
// @Success 202 {int} map
// @router /:email [put]
func (this *EmployeeSigningController) Update() {
	action := "enable/unable employee signing"
	employeeEmail := this.GetString(":email")

	pl, err := this.tokenPayloadOfCorpManager()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, action)
		return
	}

	if !pl.hasEmployee(employeeEmail) {
		this.sendFailedResponse(400, util.ErrNotSameCorp, fmt.Errorf("not same corp"), action)
		return
	}

	var info models.EmployeeSigningUdateInfo
	if err := this.fetchInputPayload(&info); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, action)
		return
	}

	err = (&info).Update(pl.LinkID, employeeEmail)
	if err != nil {
		this.sendFailedResponse(0, "", err, action)
		return
	}

	this.sendResponse("enabled employee successfully", 0)

	if info.Enabled {
		this.sendEmailToEmployee(
			pl, employeeEmail, employeeEmail,
			email.EmployeeActionActive, "Activate the CLA signing")

	} else {
		this.sendEmailToEmployee(
			pl, employeeEmail, employeeEmail,
			email.EmployeeActionInactive, "Inactivate the CLA signing")
	}
}

// @Title Delete
// @Description delete employee signing
// @Param	:email		path 	string	true		"email"
// @Success 204 {string} delete success!
// @router /:email [delete]
func (this *EmployeeSigningController) Delete() {
	action := "delete employee signing"
	employeeEmail := this.GetString(":email")

	pl, err := this.tokenPayloadOfCorpManager()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, action)
		return
	}

	if !pl.hasEmployee(employeeEmail) {
		this.sendFailedResponse(400, util.ErrNotSameCorp, fmt.Errorf("not same corp"), action)
		return
	}

	err = models.DeleteEmployeeSigning(pl.LinkID, employeeEmail)
	if err != nil {
		this.sendFailedResponse(0, "", err, action)
		return
	}

	this.sendResponse("delete employee successfully", 0)

	this.sendEmailToEmployee(
		pl, employeeEmail, employeeEmail, email.EmployeeActionRemoving, "Remove employee")
}

func (this *EmployeeSigningController) notifyManagers(managers []dbmodels.CorporationManagerListResult, info *models.EmployeeSigning, orgInfo *dbmodels.OrgInfo) {
	ms := make([]string, 0, len(managers))
	to := make([]string, 0, len(managers))
	for _, item := range managers {
		if item.Role == dbmodels.RoleManager {
			to = append(to, item.Email)
			ms = append(ms, fmt.Sprintf("%s: %s", item.Name, item.Email))
		}
	}

	msg := email.EmployeeSigning{
		Name:       info.Name,
		Org:        orgInfo.OrgAlias,
		ProjectURL: orgInfo.ProjectURL(),
		Managers:   "  " + strings.Join(ms, "\n  "),
	}
	sendEmailToIndividual(
		info.Email, orgInfo.OrgEmail,
		fmt.Sprintf("Signing CLA on project of \"%s\"", msg.Org),
		msg,
	)

	msg1 := email.NotifyingManager{
		Org:              orgInfo.OrgAlias,
		EmployeeEmail:    info.Email,
		ProjectURL:       orgInfo.ProjectURL(),
		URLOfCLAPlatform: conf.AppConfig.CLAPlatformURL,
	}
	sendEmail(to, orgInfo.OrgEmail, "An employee has signed CLA", msg1)
}

func (this *EmployeeSigningController) sendEmailToEmployee(pl *acForCorpManagerPayload, employeeName, employeeEmail, action, subject string) {
	msg := email.EmployeeNotification{
		Action:     action,
		Name:       employeeName,
		Manager:    pl.Email,
		Org:        pl.OrgAlias,
		ProjectURL: pl.ProjectURL(),
	}
	sendEmailToIndividual(employeeEmail, pl.OrgEmail, subject, msg)
}
