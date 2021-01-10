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

	claInfo, fr := getCLAInfoSigned(linkID, claLang, dbmodels.ApplyToIndividual)
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
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
		if claInfo == nil {
			this.sendFailedResponse(500, errSystemError, fmt.Errorf("no cla info, impossible"), action)
			return
		}
	}

	if claInfo.CLAHash != this.GetString(":cla_hash") {
		this.sendFailedResponse(400, errUnmatchedCLA, fmt.Errorf("invalid cla"), action)
		return
	}

	info.Info = getSingingInfo(info.Info, claInfo.Fields)

	if err := (&info).Create(linkID, false); err != nil {
		if err.IsErrorOf(models.ErrNoLinkOrResigned) {
			this.sendFailedResponse(400, errResigned, err, action)
		} else {
			this.sendModelErrorAsResp(err, action)
		}
		return
	}

	this.sendSuccessResp("sign successfully")

	this.notifyManagers(managers, &info, orgInfo)
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

	orgCLA := &models.OrgCLA{ID: pl.LinkID}
	if err := orgCLA.Get(); err != nil {
		sendResp(convertDBError1(err))
		return
	}

	linkID, fr := getLinkID(
		orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID, dbmodels.ApplyToIndividual,
	)
	if fr != nil {
		sendResp(fr)
		return
	}

	r, merr := models.ListIndividualSigning(linkID, pl.Email, this.GetString("cla_language"))
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

	corpClaOrg := &models.OrgCLA{ID: pl.LinkID}
	if err := corpClaOrg.Get(); err != nil {
		sendResp(convertDBError1(err))
		return
	}

	linkID, fr := getLinkID(
		corpClaOrg.Platform, corpClaOrg.OrgID, corpClaOrg.RepoID, dbmodels.ApplyToIndividual,
	)
	if fr != nil {
		sendResp(fr)
		return
	}

	var info models.EmployeeSigningUdateInfo
	if fr := this.fetchInputPayload(&info); fr != nil {
		sendResp(fr)
		return
	}

	if err := (&info).Update(linkID, employeeEmail); err != nil {
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

	corpClaOrg := &models.OrgCLA{ID: pl.LinkID}
	if err := corpClaOrg.Get(); err != nil {
		sendResp(convertDBError1(err))
		return
	}

	linkID, fr := getLinkID(
		corpClaOrg.Platform, corpClaOrg.OrgID, corpClaOrg.RepoID, dbmodels.ApplyToIndividual,
	)
	if fr != nil {
		sendResp(fr)
		return
	}

	if err := models.DeleteEmployeeSigning(linkID, employeeEmail); err != nil {
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
		URLOfCLAPlatform: conf.AppConfig.CLAPlatformURL,
	}
	sendEmail(to, orgInfo.OrgEmail, "An employee has signed CLA", msg1)
}
