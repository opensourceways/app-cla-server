package controllers

import (
	"fmt"
	"net/http"

	"github.com/opensourceways/app-cla-server/models"
)

type CorporationManagerController struct {
	baseController
}

func (this *CorporationManagerController) Prepare() {
	switch this.apiRequestMethod() {
	case http.MethodPut:
		// add administrator
		this.apiPrepare(PermissionOwnerOfOrg)

	case http.MethodPatch:
		// reset password of manager
		this.apiPrepareWithAC(
			&accessController{Payload: &acForCorpManagerPayload{}},
			[]string{PermissionCorpAdmin, PermissionEmployeeManager},
		)
	}
}

// @Title Put
// @Description add corporation administrator
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 202 {int} map
// @Failure util.ErrPDFHasNotUploaded
// @Failure util.ErrNumOfCorpManagersExceeded
// @router /:link_id/:email [put]
func (this *CorporationManagerController) Put() {
	action := "add corp administrator"
	linkID := this.GetString(":link_id")
	corpEmail := this.GetString(":email")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	orgInfo, merr := models.GetOrgOfLink(linkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	// call models.GetCorpSigningBasicInfo before models.IsCorpSigningPDFUploaded
	// to check wheather corp has signed
	corpSigning, merr := models.GetCorpSigningBasicInfo(linkID, corpEmail)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	uploaded, err := models.IsCorpSigningPDFUploaded(linkID, corpEmail)
	if err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}
	if !uploaded {
		this.sendFailedResponse(
			400, errUnuploaded,
			fmt.Errorf("pdf corporation signed has not been uploaded"), action)
		return
	}

	added, merr := models.CreateCorporationAdministrator(linkID, corpSigning.AdminName, corpEmail)
	if merr != nil {
		if merr.IsErrorOf(models.ErrNoLinkOrManagerExists) {
			this.sendFailedResponse(400, errCorpManagerExists, merr, action)
		} else {
			this.sendModelErrorAsResp(merr, action)
		}
		return
	}

	this.sendSuccessResp(action + " successfully")

	notifyCorpAdmin(orgInfo, added)
}

// @Title Patch
// @Description reset password of corporation manager
// @Param	body		body 	dbmodels.CorporationManagerResetPassword	true		"body for resetting password"
// @Success 204 {string} "reset password successfully"
// @Failure 401 missing_token:               token is missing
// @Failure 402 unknown_token:               token is unknown
// @Failure 403 expired_token:               token is expired
// @Failure 404 unauthorized_token:          the permission of token is unmatched
// @Failure 405 error_parsing_api_body:      parse payload of request failed
// @Failure 406 same_password:               the old and new passwords are same
// @Failure 407 too_short_or_long_password:  the length of new password is too short or long
// @Failure 408 invalid_password:            the format of new password is invalid
// @Failure 409 corp_manager_does_not_exist: manager may be removed
// @Failure 410 wrong_old_password:          the old password is not correct
// @Failure 411 frequent_operation:          don't operate frequently
// @Failure 500 system_error:                system error
// @router / [patch]
func (this *CorporationManagerController) Patch() {
	action := "reset password of corp's manager"
	sendResp := this.newFuncForSendingFailedResp(action)

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	var info models.CorporationManagerResetPassword
	if fr := this.fetchInputPayload(&info); fr != nil {
		sendResp(fr)
		return
	}

	if err := info.Validate(); err != nil {
		sendResp(parseModelError(err))
		return
	}

	if err := (&info).Reset(pl.LinkID, pl.Email); err != nil {
		if err.IsErrorOf(models.ErrNoLinkOrNoManagerOrFO) {
			this.sendFailedResponse(400, errFrequentOperation, err, action)
		} else {
			this.sendModelErrorAsResp(err, action)
		}
		return
	}

	this.sendSuccessResp("reset password successfully")
}
