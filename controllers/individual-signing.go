package controllers

import (
	"fmt"

	"github.com/astaxie/beego"

	// "github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	// "github.com/opensourceways/app-cla-server/worker"
)

type IndividualSigningController struct {
	beego.Controller
}

func (this *IndividualSigningController) Prepare() {
	if getRouterPattern(&this.Controller) == "/v1/individual-signing/:platform/:org/:repo" {
		return
	}

	apiPrepare(&this.Controller, []string{PermissionIndividualSigner}, nil)
}

// @Title Post
// @Description sign as individual
// @Param	:cla_org_id	path 	string				true		"cla org id"
// @Param	body		body 	models.IndividualSigning	true		"body for individual signing"
// @Success 201 {int} map
// @router /:cla_org_id [post]
func (this *IndividualSigningController) Post() {
	var statusCode = 201
	var errCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body)
	}()

	claOrgID, err := fetchStringParameter(&this.Controller, ":cla_org_id")
	if err != nil {
		reason = err
		errCode = ErrInvalidParameter
		statusCode = 400
		return
	}

	var info models.IndividualSigning
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = ErrInvalidParameter
		statusCode = 400
		return
	}

	/*
		_, emailCfg, err := getEmailConfig(claOrgID)
		if err != nil {
			reason = err
			statusCode = 500
			return
		}

		msg, err := email.GenIndividualSigningNotificationMsg(nil)
		if err != nil {
			reason = err
			statusCode = 500
			return
		}
		msg.To = []string{info.Email}
	*/
	if err := (&info).Create(claOrgID, true); err != nil {
		reason = fmt.Errorf("Failed to sign as individual, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	body = "sign successfully"

	// worker.GetEmailWorker().SendSimpleMessage(emailCfg, msg)
}

// @Title Check
// @Description check whether contributor has signed cla
// @Param	platform	path 	string	true		"code platform"
// @Param	org		path 	string	true		"org"
// @Param	repo		path 	string	true		"repo"
// @Param	email		query 	string	true		"email"
// @Success 200
// @router /:platform/:org/:repo [get]
func (this *IndividualSigningController) Check() {
	var statusCode = 200
	var errCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body)
	}()

	params := []string{":platform", ":org", ":repo", "email"}
	if err := checkAPIStringParameter(&this.Controller, params); err != nil {
		reason = err
		errCode = ErrInvalidParameter
		statusCode = 400
		return
	}

	v, err := models.IsIndividualSigned(
		this.GetString(":platform"),
		this.GetString(":org"),
		this.GetString(":repo"),
		this.GetString("email"),
	)
	if err != nil {
		reason = fmt.Errorf("Failed to check individual signing, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	body = map[string]bool{
		"signed": v,
	}
}
