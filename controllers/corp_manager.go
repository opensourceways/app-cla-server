package controllers

import (
	"strings"

	"github.com/opensourceways/app-cla-server/models"
)

type CorporationManagerController struct {
	baseController
}

func (ctl *CorporationManagerController) Prepare() {
	if strings.HasSuffix(ctl.routerPattern(), ":signing_id") {
		// add administrator
		ctl.apiPrepare(PermissionOwnerOfOrg)

		return
	}

	if ctl.isPostRequest() {
		// login
		return
	}

	// change password of manager or logout or get basic info
	ctl.apiPrepareWithAC(
		&accessController{Payload: &acForCorpManagerPayload{}},
		[]string{PermissionCorpAdmin, PermissionEmployeeManager},
	)
}

// @Title AddCorpAdmin
// @Description add corporation administrator
// @Tags CorpManager
// @Accept json
// @Param  link_id     path  string  true  "link id"
// @Param  signing_id  path  string  true  "signing id"
// @Success 202 {object} controllers.respData
// @Failure util.ErrPDFHasNotUploaded
// @Failure util.ErrNumOfCorpManagersExceeded
// @router /:link_id/:signing_id [post]
func (ctl *CorporationManagerController) AddCorpAdmin() {
	csId := ctl.GetString(":signing_id")
	action := "community manager adds corp admin of signing: " + csId

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	linkId, added, merr := models.CreateCorporationAdministratorByAdapter(pl.UserId, csId)
	if merr != nil {
		if merr.IsErrorOf(models.ErrNoLinkOrManagerExists) {
			ctl.sendFailedResponse(400, errCorpManagerExists, merr, action)
		} else {
			ctl.sendModelErrorAsResp(merr, action)
		}

		return
	}

	orgInfo, merr := models.GetLink(linkId)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)

		return
	}

	ctl.sendSuccessResp(action, "successfully")

	notifyCorpAdmin(linkId, &orgInfo, &added)
}

// @Title ChangePassword
// @Description corporation manager changes password
// @Tags CorpManager
// @Accept json
// @Success 202 {object} controllers.respData
// @Failure util.ErrInvalidAccountOrPw
// @router / [put]
func (ctl *CorporationManagerController) ChangePassword() {
	action := "corp admin or employee manager changes password"
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

	ctl.sendSuccessResp(action, "successfully")
}

// @Title GetBasicInfo
// @Description get basic info of corporation manager
// @Tags CorpManager
// @Accept json
// @Success 202 {object} controllers.corpManagerInfo
// @router / [get]
func (ctl *CorporationManagerController) GetBasicInfo() {
	action := "get basic info of corp admin or employee manager"
	sendResp := ctl.newFuncForSendingFailedResp(action)

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		sendResp(fr)
		return
	}

	orgInfo, merr := models.GetLink(pl.LinkID)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)

		return
	}

	v, merr := models.GetUserInfo(pl.UserId)
	if merr != nil {
		ctl.sendModelErrorAsResp(merr, action)

		return
	}

	ctl.sendSuccessResp(action, corpManagerInfo{
		Org:                 orgInfo.OrgAlias,
		LinkId:              pl.LinkID,
		CorpManagerUserInfo: v,
	})

}

type corpManagerInfo struct {
	Org    string `json:"org"`
	LinkId string `json:"link_id"`

	models.CorpManagerUserInfo
}
