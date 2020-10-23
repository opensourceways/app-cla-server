package controllers

import (
	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type VerificationCodeController struct {
	beego.Controller
}

func (this *VerificationCodeController) Prepare() {
	if getHeader(&this.Controller, headerToken) != "" {
		apiPrepare(&this.Controller, []string{PermissionIndividualSigner}, nil)
	}
}

// @Title Post
// @Description send verification code when signing
// @Param	:cla_org_id	path 	string					true		"cla org id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 201 {int} map
// @Failure util.ErrSendingEmail
// @router /:cla_org_id/:email [post]
func (this *VerificationCodeController) Post() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "send verification code")
	}()

	if err := checkAPIStringParameter(&this.Controller, []string{":cla_org_id", ":email"}); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}
	orgCLAID := this.GetString(":cla_org_id")
	individualEmail := this.GetString(":email")

	orgCLA := &models.OrgCLA{ID: orgCLAID}
	if err := orgCLA.Get(); err != nil {
		reason = err
		return
	}

	m := map[string]string{
		dbmodels.ApplyToCorporation: models.ActionCorporationSigning,
		dbmodels.ApplyToIndividual:  models.ActionEmployeeSigning,
	}

	code, err := models.CreateVerificationCode(
		individualEmail, m[orgCLA.ApplyTo],
		conf.AppConfig.VerificationCodeExpiry,
	)
	if err != nil {
		reason = err
		return
	}

	body = "create verification code successfully"

	msg := email.CorpSigningVerificationCode{Code: code}
	sendEmailToIndividual(individualEmail, orgCLA.OrgEmail, "verification code", msg)
}
