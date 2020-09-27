package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/worker"
)

type CorporationSigningController struct {
	beego.Controller
}

func (this *CorporationSigningController) Prepare() {
	method := this.Ctx.Request.Method

	if method == http.MethodGet || method == http.MethodPut {
		apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, nil)
	}
}

// @Title Post
// @Description sign as corporation
// @Param	:cla_org_id	path 	string					true		"cla org id"
// @Param	body		body 	models.CorporationSigningCreateOption	true		"body for corporation signing"
// @Success 201 {int} map
// @router /:cla_org_id [post]
func (this *CorporationSigningController) Post() {
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

	var info models.CorporationSigningCreateOption
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = ErrInvalidParameter
		statusCode = 400
		return
	}

	if err := (&info).Validate(); err != nil {
		reason = fmt.Errorf("Failed to sign as corporation, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	claOrg, emailCfg, err := getEmailConfig(claOrgID)
	if err != nil {
		reason = fmt.Errorf("Failed to sign as corporation, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	cla := &models.CLA{ID: claOrg.CLAID}
	if err := cla.Get(); err != nil {
		reason = fmt.Errorf("Failed to sign as corporation, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	if err := (&info).Create(claOrgID); err != nil {
		reason = fmt.Errorf("Failed to sign as corporation, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	body = "sign successfully"

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(claOrg, &info.CorporationSigning, cla, emailCfg)
}

// @Title GetAll
// @Description get all the corporations which have signed to a org
// @router / [get]
func (this *CorporationSigningController) GetAll() {
	var statusCode = 200
	var errCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body)
	}()

	opt := models.CorporationSigningListOption{
		Platform:    this.GetString("platform"),
		OrgID:       this.GetString("org_id"),
		RepoID:      this.GetString("repo_id"),
		CLALanguage: this.GetString("cla_language"),
	}

	r, err := opt.List()
	if err != nil {
		reason = fmt.Errorf("Failed to list corporation, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
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
		sendResponse1(&this.Controller, statusCode, reason, body)
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

// @Title SendVerifiCode
// @Description send verification code when signing as Corporation
// @Param	:cla_org_id	path 	string					true		"cla org id"
// @Param	body		body 	models.CorporationSigningVerifCode	true		"body for sending verification code"
// @Success 201 {int} map
// @router /verifi-code/:cla_org_id [post]
func (this *CorporationSigningController) SendVerifiCode() {
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

	var info models.CorporationSigningVerifCode
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = ErrInvalidParameter
		statusCode = 400
		return
	}

	_, emailCfg, err := getEmailConfig(claOrgID)
	if err != nil {
		reason = fmt.Errorf("Failed to send verification code, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	ec, err := email.GetEmailClient(emailCfg.Platform)
	if err != nil {
		reason = fmt.Errorf("Failed to send verification code, err:%s", err.Error())
		errCode = ErrUnknownEmailPlatform
		statusCode = 500
		return
	}

	code, err := info.Create(conf.AppConfig.VerificationCodeExpiry)
	if err != nil {
		reason = fmt.Errorf("Failed to send verification code, err:%s", err.Error())
		statusCode, errCode = convertDBError(err)
		return
	}

	msg := email.EmailMessage{
		To:      []string{info.Email},
		Content: code,
		Subject: "verification code",
	}
	if err := ec.SendEmail(emailCfg.Token, &msg); err != nil {
		reason = fmt.Errorf("Failed to send verification code, err: %s", err.Error())
		errCode = ErrSendingEmail
		statusCode = 500
		return
	}

	body = "verification code has been sent successfully"
}
