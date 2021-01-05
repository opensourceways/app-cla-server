package controllers

import (
	"fmt"
	"net/http"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
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
			[]string{PermissionCorporAdmin, PermissionEmployeeManager},
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
// @router /:org_cla_id/:email [put]
func (this *CorporationManagerController) Put() {
	action := "add corp administrator"
	sendResp := this.newFuncForSendingFailedResp(action)
	orgCLAID := this.GetString(":org_cla_id")
	corpEmail := this.GetString(":email")

	orgCLA, statusCode, errCode, reason := canAccessOrgCLA(&this.Controller, orgCLAID)
	if reason != nil {
		this.sendFailedResponse(statusCode, errCode, reason, action)
		return
	}

	uploaded, err := models.IsCorpSigningPDFUploaded(orgCLAID, corpEmail)
	if err != nil {
		sendResp(convertDBError1(err))
		return
	}
	if !uploaded {
		this.sendFailedResponse(
			400, util.ErrPDFHasNotUploaded,
			fmt.Errorf("pdf corporation signed has not been uploaded"), action)
		return
	}

	_, corpSigning, err := models.GetCorporationSigningDetail(
		orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID, corpEmail,
	)
	if err != nil {
		sendResp(convertDBError1(err))
		return
	}

	added, err := models.CreateCorporationAdministrator(orgCLAID, corpSigning.AdminName, corpEmail)
	if err != nil {
		sendResp(convertDBError1(err))
		return
	}

	this.sendSuccessResp("add manager successfully")

	notifyCorpManagerWhenAdding(orgCLA, added)
}

// @Title Patch
// @Description reset password of corporation administrator
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

	if errCode, reason := info.Validate(); reason != nil {
		this.sendFailedResponse(400, errCode, reason, action)
		return
	}

	if err := (&info).Reset(pl.OrgCLAID, pl.Email); err != nil {
		sendResp(convertDBError1(err))
		return
	}

	this.sendSuccessResp("reset password successfully")
}
