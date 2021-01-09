package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/models"
)

type CLAController struct {
	baseController
}

func (this *CLAController) Prepare() {
	apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg})
}

// @Title Delete CLA
// @Description delete cla
// @Param	uid		path 	string	true		"cla id"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:uid [delete]
func (this *CLAController) Delete() {
	var statusCode = 0
	var reason error
	var body string

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	uid := this.GetString(":uid")
	if uid == "" {
		reason = fmt.Errorf("missing cla id")
		statusCode = 400
		return
	}

	cla := models.CLA{ID: uid}

	if err := (&cla).Delete(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "delete cla successfully"
}

// @Title Get
// @Description get cla by uid
// @Param	uid		path 	string	true		"The key for cla"
// @Success 200 {object} models.CLA
// @Failure 403 :uid is empty
// @router /:uid [get]
func (this *CLAController) Get() {
	var statusCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	uid := this.GetString(":uid")
	if uid == "" {
		reason = fmt.Errorf("missing cla id")
		statusCode = 400
		return
	}

	cla := models.CLA{ID: uid}

	if err := (&cla).Get(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = cla
}

// @Title List
// @Description list clas of link
// @Param	link_id		path 	string	true		"link id"
// @Success 200 {string} delete success!
// @Failure 403 uid is empty
// @router /:link_id [get]
func (this *CLAController) List() {
	doWhat := "delete cla"
	linkID := this.GetString(":link_id")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, doWhat)
		return
	}

	clas, merr := models.GetAllCLA(linkID)
	if merr != nil {
		this.sendModelErrorAsResp(merr, doWhat)
		return
	}

	this.sendSuccessResp(clas)
}
