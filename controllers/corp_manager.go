package controllers

import (
	"net/http"
	"strings"

	"github.com/opensourceways/app-cla-server/models"
)

type CorporationManagerController struct {
	baseController
}

func (ctl *CorporationManagerController) Prepare() {
	m := ctl.apiRequestMethod()

	if m == http.MethodPost {
		// login
		return
	}

	if m == http.MethodPut && strings.HasSuffix(ctl.routerPattern(), ":signing_id") {
		// add administrator
		ctl.apiPrepare(PermissionOwnerOfOrg)

		return
	}

	// change password of manager or logout
	ctl.apiPrepareWithAC(
		&accessController{Payload: &acForCorpManagerPayload{}},
		[]string{PermissionCorpAdmin, PermissionEmployeeManager},
	)
}

// @Title Put
// @Description add corporation administrator
// @Tags CorpManager
// @Accept json
// @Param  link_id     path  string  true  "link id"
// @Param  signing_id  path  string  true  "signing id"
// @Success 202 {int} map
// @Failure util.ErrPDFHasNotUploaded
// @Failure util.ErrNumOfCorpManagersExceeded
// @router /:link_id/:signing_id [put]
func (ctl *CorporationManagerController) Put() {
	action := "add corp administrator"
	linkID := ctl.GetString(":link_id")
	csId := ctl.GetString(":signing_id")

	pl, fr := ctl.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	orgInfo, merr := models.GetLink(linkID)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)

		return
	}

	added, merr := models.CreateCorporationAdministratorByAdapter(csId)
	if merr != nil {
		if merr.IsErrorOf(models.ErrNoLinkOrManagerExists) {
			ctl.sendFailedResponse(400, errCorpManagerExists, merr, action)
		} else {
			ctl.sendModelErrorAsResp(merr, action)
		}
		return
	}

	ctl.sendSuccessResp(action + " successfully")

	notifyCorpAdmin(&orgInfo, &added)
}

// @Title Patch
// @Description change password of corporation manager
// @Tags CorpManager
// @Accept json
// @Success 204 {int} map
// @Failure util.ErrInvalidAccountOrPw
// @router / [patch]
func (ctl *CorporationManagerController) Patch() {
	action := "change password of corp's manager"
	sendResp := ctl.newFuncForSendingFailedResp(action)

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	var info models.CorporationManagerChangePassword
	if fr := ctl.fetchInputPayload(&info); fr != nil {
		sendResp(fr)
		return
	}

	if err := models.ChangePassword(pl.UserId, &info); err != nil {
		ctl.sendModelErrorAsResp(err, action)
		return
	}

	ctl.logout()

	ctl.sendSuccessResp("change password successfully")
}
