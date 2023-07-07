package controllers

import (
	"net/http"
	"strings"

	"github.com/opensourceways/app-cla-server/models"
)

type CorporationManagerController struct {
	baseController
}

func (this *CorporationManagerController) Prepare() {
	m := this.apiRequestMethod()

	if m == http.MethodPost {
		return
	}

	if m == http.MethodPut && strings.HasSuffix(this.routerPattern(), ":signing_id") {
		// add administrator
		this.apiPrepare(PermissionOwnerOfOrg)

		return
	}

	// change password of manager
	this.apiPrepareWithAC(
		&accessController{Payload: &acForCorpManagerPayload{}},
		[]string{PermissionCorpAdmin, PermissionEmployeeManager},
	)
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

	orgInfo, merr := models.GetLink(linkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)

		return
	}

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

	notifyCorpAdmin(&orgInfo, &added)
}

// @Title Patch
// @Description change password of corporation manager
// @Success 204 {int} map
// @Failure util.ErrInvalidAccountOrPw
// @router / [patch]
func (this *CorporationManagerController) Patch() {
	action := "change password of corp's manager"
	sendResp := this.newFuncForSendingFailedResp(action)

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	var info models.CorporationManagerChangePassword
	if fr := this.fetchInputPayload(&info); fr != nil {
		sendResp(fr)
		return
	}

	if err := models.ChangePassword(pl.UserId, &info); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("change password successfully")
}
