package controllers

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
)

type EmployeeSigningController struct {
	baseController
}

func (this *EmployeeSigningController) Prepare() {
	this.stopRunIfSignSerivceIsUnabled()

	if this.isPostRequest() {
		// sign as employee
		this.apiPrepare("")
	} else {
		// get, update and delete employee
		this.apiPrepare(PermissionEmployeeManager)
	}
}

// @Title Post
// @Description sign employee cla
// @Param	:link_id	path 	string				true		"link id"
// @Param	:cla_lang	path 	string				true		"cla language"
// @Param	:cla_hash	path 	string				true		"the hash of cla content"
// @Param	body		body 	models.EmployeeSigning		true		"body for individual signing"
// @Success 201 {string} "sign successfully"
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 405 error_parsing_api_body:     parse payload of request failed
// @Failure 406 unmatched_email:            the email is not same as the one which signer sets on the code platform
// @Failure 407 unmatched_user_id:          the user id is not same as the one which was fetched from code platform
// @Failure 408 expired_verification_code:  the verification code is expired
// @Failure 409 wrong_verification_code:    the verification code is wrong
// @Failure 410 no_link:                    the link id is not exists
// @Failure 411 no_employee_manager:        there is not any employee managers for the corresponding corp
// @Failure 412 unmatched_cla:              the cla hash is not equal to the one of backend server
// @Failure 413 resigned:                   the signer has signed the cla
// @Failure 500 system_error:               system error
// @router /:link_id/:cla_lang/:cla_id [post]
func (this *EmployeeSigningController) Post() {
	action := "sign employeee cla"
	linkID := this.GetString(":link_id")
	claLang := this.GetString(":cla_lang")
	claId := this.GetString(":cla_id")

	var info models.EmployeeSigning
	if fr := this.fetchInputPayload(&info); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	info.CLALanguage = claLang

	if err := info.Validate(linkID); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	orgInfo, claInfo, merr := models.GetLinkCLA(linkID, claId)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	info.Info = getSingingInfo(info.Info, claInfo.Fields)
	info.CLAId = claId

	managers, merr := models.SignEmployeeCLA(&info)
	if merr != nil {
		if merr.IsErrorOf(models.ErrNoLinkOrResigned) {
			this.sendFailedResponse(400, errResigned, merr, action)
		} else {
			this.sendModelErrorAsResp(merr, action)
		}

		return
	}

	this.sendSuccessResp("sign successfully")
	this.notifyManagers(managers, &info, &orgInfo)
}

// @Title GetAll
// @Description get all the employees
// @Success 200 {object} dbmodels.IndividualSigningBasicInfo
// @Failure 400 missing_token:      token is missing
// @Failure 401 unknown_token:      token is unknown
// @Failure 402 expired_token:      token is expired
// @Failure 403 unauthorized_token: the permission of token is unmatched
// @Failure 500 system_error:       system error
// @router / [get]
func (this *EmployeeSigningController) GetAll() {
	action := "list employees"

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	r, merr := models.ListEmployeeSignings(pl.SigningId)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendSuccessResp(r)
}

// @Title Update
// @Description enable/unable employee signing
// @Param  signing_id  path  string                           true  "employee signing id"
// @Param  param       body  models.EmployeeSigningUdateInfo  true  "body of updating employee signing"
// @Success 202 {int} map
// @router /:signing_id [put]
func (this *EmployeeSigningController) Update() {
	action := "enable/unable employee signing"
	sendResp := this.newFuncForSendingFailedResp(action)
	employeeSigningId := this.GetString(":signing_id")

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	orgInfo, merr := models.GetLink(pl.LinkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)

		return
	}

	var info models.EmployeeSigningUdateInfo
	if fr := this.fetchInputPayload(&info); fr != nil {
		sendResp(fr)
		return
	}

	employeeEmail, err := models.UpdateEmployeeSigning(
		pl.SigningId, employeeSigningId, info.Enabled,
	)
	if err != nil {
		if err.IsErrorOf(models.ErrNoLinkOrUnsigned) {
			this.sendFailedResponse(400, errUnsigned, err, action)
		} else {
			this.sendModelErrorAsResp(err, action)
		}
		return
	}

	this.sendSuccessResp("enabled employee successfully")

	msg := this.newEmployeeNotification(employeeEmail, &orgInfo, pl.Email)
	if info.Enabled {
		msg.Active = true
		sendEmailToIndividual(employeeEmail, &orgInfo, "Activate CLA signing", msg)
	} else {
		msg.Inactive = true
		sendEmailToIndividual(employeeEmail, &orgInfo, "Inactivate CLA signing", msg)
	}
}

// @Title Delete
// @Description delete employee signing
// @Param  signing_id  path  string  true  "employee signing id"
// @Success 204 {string} delete success!
// @router /:signing_id [delete]
func (this *EmployeeSigningController) Delete() {
	action := "delete employee signing"
	employeeSigningId := this.GetString(":signing_id")

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	orgInfo, merr := models.GetLink(pl.LinkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)

		return
	}

	employeeEmail, err := models.RemoveEmployeeSigning(
		pl.SigningId, employeeSigningId,
	)
	if err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("delete employee successfully")

	msg := this.newEmployeeNotification(employeeEmail, &orgInfo, pl.Email)
	msg.Removing = true
	sendEmailToIndividual(employeeEmail, &orgInfo, "Remove employee", msg)
}

func (this *EmployeeSigningController) notifyManagers(managers []dbmodels.CorporationManagerListResult, info *models.EmployeeSigning, orgInfo *models.OrgInfo) {
	ms := make([]string, 0, len(managers))
	to := make([]string, 0, len(managers))
	for _, item := range managers {
		to = append(to, item.Email)
		ms = append(ms, fmt.Sprintf("%s: %s", item.Name, item.Email))
	}

	msg := emailtmpl.EmployeeSigning{
		Name:       info.Name,
		Org:        orgInfo.OrgAlias,
		ProjectURL: orgInfo.ProjectURL(),
		Managers:   "  " + strings.Join(ms, "\n  "),
	}
	sendEmailToIndividual(
		info.Email, orgInfo,
		fmt.Sprintf("Signing CLA on project of \"%s\"", msg.Org),
		msg,
	)

	msg1 := emailtmpl.NotifyingManager{
		Org:              orgInfo.OrgAlias,
		EmployeeEmail:    info.Email,
		ProjectURL:       orgInfo.ProjectURL(),
		URLOfCLAPlatform: config.AppConfig.CLAPlatformURL,
	}
	sendEmail(to, orgInfo, "An employee has signed CLA", msg1)
}

func (this *EmployeeSigningController) newEmployeeNotification(
	employeeName string, orgInfo *models.OrgInfo, managerEmail string,
) *emailtmpl.EmployeeNotification {
	return &emailtmpl.EmployeeNotification{
		Name:       employeeName,
		Manager:    managerEmail,
		Org:        orgInfo.OrgAlias,
		ProjectURL: orgInfo.ProjectURL(),
	}
}
