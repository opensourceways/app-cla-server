package controllers

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
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
		if strings.HasSuffix(this.routerPattern(), "/:link_id/:email") {
			this.apiPrepare(PermissionOwnerOfOrg)
		} else {
			// get, update and delete employee
			this.apiPrepare(PermissionEmployeeManager)
		}
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
// @router /:link_id/:cla_lang/:cla_hash [post]
func (this *EmployeeSigningController) Post() {
	action := "sign employeee cla"
	linkID := this.GetString(":link_id")
	claLang := this.GetString(":cla_lang")

	var info models.EmployeeSigning
	if fr := this.fetchInputPayload(&info); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	info.CLALanguage = claLang

	if err := (&info).Validate(linkID); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	orgInfo, merr := models.GetOrgOfLink(linkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	detail, fr := getCorporationDetail(
		models.SigningIndex{
			LinkId:    linkID,
			SigningId: info.CorpSigningId,
		},
	)
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if len(detail.Managers) <= 0 {
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
		this.notifyManagers(detail.Managers, &info, orgInfo)
	}
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

	r, merr := models.ListEmployeeSigning(
		pl.signingIndex(), this.GetString("cla_language"),
	)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	this.sendSuccessResp(r)
}

// @Title List
// @Description get all the employees by community manager
// @Param	:link_id	path 	string		true		"link id"
// @Param	:signing_id	path 	string		true		"signing id"
// @Success 200 {object} dbmodels.IndividualSigningBasicInfo
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 405 unknown_link:               unkown link id
// @Failure 406 not_yours_org:              the link doesn't belong to your community
// @Failure 500 system_error:               system error
// @router /:link_id/:signing_id [get]
func (this *EmployeeSigningController) List() {
	action := "list employees"
	index := genSigningIndex(&this.Controller)

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if fr := pl.isOwnerOfLink(index.LinkId); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	r, merr := models.ListEmployeeSigning(index, "")
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

	detail, fr := getCorporationDetail(pl.signingIndex())
	if fr != nil {
		fr.statusCode = 500
		sendResp(fr)
		return
	}

	if !detail.HasDomain(util.EmailSuffix(employeeEmail)) {
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

	detail, fr := getCorporationDetail(pl.signingIndex())
	if fr != nil {
		fr.statusCode = 500
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if !detail.HasDomain(util.EmailSuffix(employeeEmail)) {
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

func (this *EmployeeSigningController) notifyManagers(
	managers []dbmodels.CorporationManagerListResult,
	info *models.EmployeeSigning,
	orgInfo *models.OrgInfo,
) {
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

func (this *EmployeeSigningController) newEmployeeNotification(
	pl *acForCorpManagerPayload, employeeName string,
) *email.EmployeeNotification {
	return &email.EmployeeNotification{
		Name:       employeeName,
		Manager:    pl.Email,
		Org:        pl.OrgAlias,
		ProjectURL: pl.ProjectURL(),
	}
}
