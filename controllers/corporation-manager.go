package controllers

import (
	"net/http"

	"github.com/opensourceways/app-cla-server/models"
)

type CorporationManagerController struct {
	baseController
}

func (this *CorporationManagerController) Prepare() {
	this.stopRunIfSignSerivceIsUnabled()

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
// @Param  link_id     path  string  true  "link id"
// @Param  signing_id  path  string  true  "signing id"
// @Success 202 {int} map
// @Failure util.ErrPDFHasNotUploaded
// @Failure util.ErrNumOfCorpManagersExceeded
// @router /:link_id/:signing_id [put]
func (this *CorporationManagerController) Put() {
	action := "add corp administrator"
	linkID := this.GetString(":link_id")
	csId := this.GetString(":signing_id")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	orgInfo := pl.orgInfo(linkID)

	// lock to avoid the conflict with the deleting corp signing
	unlock, fr := lockOnRepo(orgInfo)
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	defer unlock()

	added, merr := models.CreateCorporationAdministratorByAdapter(csId)
	if merr != nil {
		if merr.IsErrorOf(models.ErrNoLinkOrManagerExists) {
			this.sendFailedResponse(400, errCorpManagerExists, merr, action)
		} else {
			this.sendModelErrorAsResp(merr, action)
		}
		return
	}

	this.sendSuccessResp(action + " successfully")

	notifyCorpAdmin(orgInfo, &added)
}

// @Title Patch
// @Description reset password of corporation manager
// @Success 204 {int} map
// @Failure util.ErrInvalidAccountOrPw
// @router / [patch]
func (this *CorporationManagerController) Patch() {
	action := "reset password of corp's manager"
	sendResp := this.newFuncForSendingFailedResp(action)

	_, fr := this.tokenPayloadBasedOnCorpManager()
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

	// TODO user index
	if err := info.ChangePassword(""); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("reset password successfully")
}
