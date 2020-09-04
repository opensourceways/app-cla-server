package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/email"
	"github.com/zengchen1024/cla-server/models"
)

type CorporationSigningController struct {
	beego.Controller
}

func (this *CorporationSigningController) Prepare() {
	method := this.Ctx.Request.Method

	if method == http.MethodGet || method == http.MethodPut {
		apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg})
	}
}

// @Title Corporation signing
// @Description sign as corporation
// @Param	body		body 	models.CorporationSigningCreateOption	true		"body for corporation signing"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router / [post]
func (this *CorporationSigningController) Post() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.CorporationSigningCreateOption
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if err := (&info).Validate(); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if err := (&info).Create(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "sign successfully"
}

// @Title GetAll
// @Description get all the corporations which have signed to a org
// @Success 200 {object} dbmodels.CorporationSigningInfo
// @router / [get]
func (this *CorporationSigningController) GetAll() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	opt := models.CorporationSigningListOption{
		Platform:    this.GetString("platform"),
		OrgID:       this.GetString("org_id"),
		RepoID:      this.GetString("repo_id"),
		CLALanguage: this.GetString("cla_language"),
	}

	r, err := opt.List()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = r
}

// @Title Enable corporation signing
// @Description enable corporation
// @Param	body		body 	models.CorporationSigning	true		"body for corporation signing"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router / [put]
func (this *CorporationSigningController) Update() {
	var statusCode = 202
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.CorporationSigningUdateInfo
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if err := (&info).Update(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "enabled corporation successfully"
}

// @Title send verification code when signing as Corporation
// @Description send verification code
// @Param	body		body 	models.CorporationSigningVerifCode	true		"body for sending verification code"
// @Success 201 {int} map
// @Failure 403 body is empty
// @router /verifi-code [post]
func (this *CorporationSigningController) SendVerifiCode() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var info models.CorporationSigningVerifCode
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &info); err != nil {
		reason = err
		statusCode = 400
		return
	}

	claOrg := &models.CLAOrg{ID: info.CLAOrgID}
	if err := claOrg.Get(); err != nil {
		reason = err
		statusCode = 400
		return
	}

	emailCfg := &models.OrgEmail{Email: claOrg.OrgEmail}
	if err := emailCfg.Get(); err != nil {
		reason = err
		statusCode = 400
		return
	}

	ec, err := email.GetEmailClient(emailCfg.Platform)
	if err != nil {
		reason = fmt.Errorf("Failtd to get email client: %s", err.Error())
		statusCode = 500
		return
	}

	expiry, err := beego.AppConfig.Int64("verification_vode_expiry")
	if err != nil {
		reason = err
		statusCode = 400
		return
	}

	code, err := info.Create(expiry)
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	msg := email.EmailMessage{
		To:      info.Email,
		Content: code,
		Subject: "verification code",
	}
	if err := ec.SendEmail(*emailCfg.Token, msg); err != nil {
		reason = fmt.Errorf("Failed to send verification code by email: %s", err.Error())
		statusCode = 500
		return
	}

	body = "verification code has been sent successfully"
}
