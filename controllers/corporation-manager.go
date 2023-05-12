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
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 202 {int} map
// @Failure util.ErrPDFHasNotUploaded
// @Failure util.ErrNumOfCorpManagersExceeded
// @router /:link_id/:signing_id [put]
func (this *CorporationManagerController) Put() {
	action := "add corp administrator"
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
	orgInfo := pl.orgInfo(index.LinkId)

	// lock to avoid the conflict with the deleting corp signing
	unlock, fr := lockOnRepo(orgInfo)
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	defer unlock()

	// call models.GetCorpSigningBasicInfo before models.IsCorpSigningPDFUploaded
	// to check wheather corp has signed
	corpSigning, merr := models.GetCorpSigningBasicInfo(&index)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	uploaded, err := models.IsCorpSigningPDFUploaded(index)
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

	added, merr := models.CreateCorporationAdministrator(index, corpSigning)
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
// @Success 204 {int} map
// @Failure util.ErrInvalidAccountOrPw
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
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("reset password successfully")
}
