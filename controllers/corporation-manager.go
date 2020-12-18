package controllers

import (
	"fmt"
	"net/http"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type CorporationManagerController struct {
	baseController
}

func (this *CorporationManagerController) Prepare() {
	switch this.getRequestMethod() {
	case http.MethodPut:
		// add administrator
		this.apiPrepare(PermissionOwnerOfOrg)

	case http.MethodPatch:
		// reset password of manager
		this.apiPrepareForSettingPW()
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
	doWhat := "add corp administrator"

	linkID := this.GetString(":link_id")
	corpEmail := this.GetString(":email")

	pl, err := this.tokenPayloadOfCodePlatform()
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}

	uploaded, err := models.IsCorpSigningPDFUploaded(linkID, corpEmail)
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}
	if !uploaded {
		err = fmt.Errorf("pdf corporation signed has not been uploaded")
		this.sendFailedResponse(400, util.ErrPDFHasNotUploaded, err, doWhat)
		return
	}

	corpSigning, err := models.GetCorporationSigningBasicInfo(linkID, corpEmail)
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	added, err := models.CreateCorporationAdministrator(linkID, corpSigning.AdminName, corpEmail)
	if err != nil {
		this.sendFailedResponse(0, "", err, doWhat)
		return
	}

	this.sendResponse("add manager successfully", 0)

	notifyCorpManagerWhenAdding(pl.orgInfo(linkID), added)
}

// @Title Patch
// @Description reset password of corporation administrator
// @Success 204 {int} map
// @Failure util.ErrInvalidAccountOrPw
// @router / [patch]
func (this *CorporationManagerController) Patch() {
	doWhat := "reset password of corp's manager"

	pl, err := this.tokenPayloadOfCorpManager()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, doWhat)
		return
	}

	info := &models.CorporationManagerResetPassword{}
	if err := this.fetchInputPayload(info); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, doWhat)
		return
	}

	if merr := info.Validate(); merr != nil {
		this.sendModelErrorAsResp(merr, doWhat)
		return
	}

	if merr := info.Reset(pl.LinkID, pl.Email); merr != nil {
		this.sendModelErrorAsResp(merr, doWhat)
		return
	}

	this.sendResponse("reset password successfully", 0)
}

func (this *CorporationManagerController) apiPrepareForSettingPW() {
	if err := this.checkPathParameter(); err != nil {
		this.sendFailedResponse(400, util.ErrInvalidParameter, err, "")
		this.StopRun()
	}

	if v := this.checkApiReqTokenForSettingPW(); v != nil {
		this.sendFailedResponse(v.statusCode, v.errCode, v.reason, "")
		this.StopRun()
	}
}

func (this *CorporationManagerController) checkApiReqTokenForSettingPW() *failedResult {
	token := this.apiReqHeader(headerToken)
	if token == "" {
		return &failedResult{
			statusCode: 401,
			errCode:    util.ErrMissingToken,
			reason:     fmt.Errorf("no token passed"),
		}
	}

	ac := &accessController{Payload: &acForCorpManagerPayload{}}
	if err := ac.ParseToken(token, conf.AppConfig.APITokenKey); err != nil {
		return &failedResult{
			statusCode: 401,
			errCode:    util.ErrUnknownToken,
			reason:     err,
		}
	}

	if err := ac.Verify([]string{PermissionCorporAdmin, PermissionEmployeeManager}); err != nil {
		return &failedResult{
			statusCode: 403,
			errCode:    util.ErrInvalidToken,
			reason:     err,
		}
	}

	this.Data[apiAccessController] = *ac
	return nil
}
