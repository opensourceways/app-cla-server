package controllers

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
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
// @router /:link_id/:cla_lang/:cla_hash [post]
func (this *EmployeeSigningController) Post() {
	action := "sign employeee cla"
	linkID := this.GetString(":link_id")
	claLang := this.GetString(":cla_lang")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	var info models.EmployeeSigning
	if fr := this.fetchInputPayload(&info); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	info.CLALanguage = claLang

	if err := (&info).Validate(linkID, pl.Email); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	orgInfo, merr := models.GetOrgOfLink(linkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	managers, merr := models.ListCorporationManagers(linkID, info.Email, dbmodels.RoleManager)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}
	if len(managers) <= 0 {
		this.sendFailedResponse(400, errNoCorpEmployeeManager, fmt.Errorf("no managers"), action)
		return
	}

	fr = signHelper(
		linkID, claLang, dbmodels.ApplyToIndividual,
		func(claInfo *models.CLAInfo) *failedApiResult {
			if claInfo.CLAHash != this.GetString(":cla_hash") {
				return newFailedApiResult(400, errUnmatchedCLA, fmt.Errorf("invalid cla"))
			}

			info.Info = getSingingInfo(info.Info, claInfo.Fields)

			if err := (&info).Create(linkID, false); err != nil {
				if err.IsErrorOf(models.ErrNoLinkOrResigned) {
					return newFailedApiResult(400, errResigned, err)
				}
				return parseModelError(err)
			}
			return nil
		},
	)
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
	} else {
		this.sendSuccessResp("sign successfully")
	}

	this.notifyManagers(managers, &info, orgInfo)
}

// @Title GetAll
// @Description get all the employees
// @Success 200 {int} map
// @router / [get]
func (this *EmployeeSigningController) GetAll() {
	action := "list employees"

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	r, merr := models.ListIndividualSigning(pl.LinkID, pl.Email, this.GetString("cla_language"))
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
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
	action := "enable/unable employee signing"
	sendResp := this.newFuncForSendingFailedResp(action)
	employeeEmail := this.GetString(":email")

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	if !pl.hasEmployee(employeeEmail) {
		this.sendFailedResponse(400, errNotSameCorp, fmt.Errorf("not same corp"), action)
		return
	}

	var info models.EmployeeSigningUdateInfo
	if fr := this.fetchInputPayload(&info); fr != nil {
		sendResp(fr)
		return
	}

	if err := (&info).Update(pl.LinkID, employeeEmail); err != nil {
		if err.IsErrorOf(models.ErrNoLinkOrUnsigned) {
			this.sendFailedResponse(400, errUnsigned, err, action)
		} else {
			this.sendModelErrorAsResp(err, action)
		}
		return
	}

	this.sendSuccessResp("enabled employee successfully")

	msg := this.newEmployeeNotification(pl, employeeEmail)
	if info.Enabled {
		msg.Active = true
		sendEmailToIndividual(employeeEmail, pl.OrgEmail, "Activate CLA signing", msg)
	} else {
		msg.Inactive = true
		sendEmailToIndividual(employeeEmail, pl.OrgEmail, "Inactivate CLA signing", msg)
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

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if !pl.hasEmployee(employeeEmail) {
		this.sendFailedResponse(400, errNotSameCorp, fmt.Errorf("not same corp"), action)
		return
	}

	if err := models.DeleteEmployeeSigning(pl.LinkID, employeeEmail); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("delete employee successfully")

	msg := this.newEmployeeNotification(pl, employeeEmail)
	msg.Removing = true
	sendEmailToIndividual(employeeEmail, pl.OrgEmail, "Remove employee", msg)
}

func (this *EmployeeSigningController) notifyManagers(managers []dbmodels.CorporationManagerListResult, info *models.EmployeeSigning, orgInfo *models.OrgInfo) {
	ms := make([]string, 0, len(managers))
	to := make([]string, 0, len(managers))
	for _, item := range managers {
		to = append(to, item.Email)
		ms = append(ms, fmt.Sprintf("%s: %s", item.Name, item.Email))
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
		URLOfCLAPlatform: config.AppConfig.CLAPlatformURL,
	}
	sendEmail(to, orgInfo.OrgEmail, "An employee has signed CLA", msg1)
}

func (this *EmployeeSigningController) newEmployeeNotification(pl *acForCorpManagerPayload, employeeName string) *email.EmployeeNotification {
	return &email.EmployeeNotification{
		Name:       employeeName,
		Manager:    pl.Email,
		Org:        pl.OrgAlias,
		ProjectURL: pl.ProjectURL(),
	}
}
