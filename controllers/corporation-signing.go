package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
	"github.com/opensourceways/app-cla-server/worker"
)

type CorporationSigningController struct {
	beego.Controller
}

func (this *CorporationSigningController) Prepare() {
	method := getRequestMethod(&this.Controller)

	if getRouterPattern(&this.Controller) == "/v1/corporation-signing/:cla_org_id/:email" {
		switch method {
		// upload pdf
		case http.MethodPatch:
			apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, nil)

		// download pdf
		case http.MethodGet:
			apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg, PermissionCorporAdmin}, nil)
		}

	} else {
		// list corp signings
		if method == http.MethodGet {
			apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, nil)
		}
	}
}

// @Title Post
// @Description sign as corporation
// @Param	:cla_org_id	path 	string					true		"cla org id"
// @Param	body		body 	models.CorporationSigningCreateOption	true		"body for corporation signing"
// @Success 201 {int} map
// @Failure util.ErrHasSigned
// @Failure util.ErrWrongVerificationCode
// @Failure util.ErrVerificationCodeExpired
// @router /:cla_org_id [post]
func (this *CorporationSigningController) Post() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "sign as corporation")
	}()

	claOrgID, err := fetchStringParameter(&this.Controller, ":cla_org_id")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	var info models.CorporationSigningCreateOption
	if err := fetchInputPayload(&this.Controller, &info); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	if err := (&info).Validate(); err != nil {
		reason = err
		return
	}

	claOrg := &models.CLAOrg{ID: claOrgID}
	if err := claOrg.Get(); err != nil {
		reason = err
		return
	}

	emailCfg := &models.OrgEmail{Email: claOrg.OrgEmail}
	if err := emailCfg.Get(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	cla := &models.CLA{ID: claOrg.CLAID}
	if err := cla.Get(); err != nil {
		reason = err
		return
	}

	err = (&info).Create(claOrgID, claOrg.Platform, claOrg.OrgID, claOrg.RepoID)
	if err != nil {
		reason = err
		return
	}

	body = "sign successfully"

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(claOrg, &info.CorporationSigning, cla, emailCfg)
}

// @Title GetAll
// @Description get all the corporations which have signed to a org
// @router / [get]
func (this *CorporationSigningController) GetAll() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "list corporation")
	}()

	opt := models.CorporationSigningListOption{
		Platform:    this.GetString("platform"),
		OrgID:       this.GetString("org_id"),
		RepoID:      this.GetString("repo_id"),
		CLALanguage: this.GetString("cla_language"),
	}

	// TODO: check whether can do this

	r, err := opt.List()
	if err != nil {
		reason = err
		statusCode, errCode = convertDBError(err)
		return
	}

	body = r
}

// @Title Upload
// @Description upload pdf of corporation signing
// @Param	:cla_org_id	path 	string					true		"cla org id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 204 {int} map
// @router /:cla_org_id/:email [patch]
func (this *CorporationSigningController) Upload() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "upload corp's signing pdf")
	}()

	// TODO: is this cla bound by the org

	if err := checkAPIStringParameter(&this.Controller, []string{":cla_org_id", ":email"}); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	f, _, err := this.GetFile("pdf")
	if err != nil {
		reason = fmt.Errorf("missing pdf file")
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		reason = err
		statusCode = 400
		return
	}

	err = models.UploadCorporationSigningPDF(this.GetString(":cla_org_id"), this.GetString(":email"), data)
	if err != nil {
		reason = err
		statusCode, errCode = convertDBError(err)
		return
	}

	body = "upload pdf of signature page successfully"
}

// @Title Download
// @Description download pdf of corporation signing
// @Param	:cla_org_id	path 	string					true		"cla org id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 200 {int} map
// @router /:cla_org_id/:email [get]
func (this *CorporationSigningController) Download() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "download corp's signing pdf")
	}()

	if err := checkAPIStringParameter(&this.Controller, []string{":cla_org_id", ":email"}); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	pdf, err := models.DownloadCorporationSigningPDF(this.GetString(":cla_org_id"), this.GetString(":email"))
	if err != nil {
		reason = err
		statusCode, errCode = convertDBError(err)
		return
	}
	if pdf == nil {
		reason = fmt.Errorf("no pdf found")
		statusCode = 500
		return
	}

	body = map[string]interface{}{
		"pdf": pdf,
	}
}

// @Title SendVerifiCode
// @Description send verification code when signing as Corporation
// @Param	:cla_org_id	path 	string					true		"cla org id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 202 {int} map
// @Failure util.ErrSendingEmail
// @router /:cla_org_id/:email [put]
func (this *CorporationSigningController) SendVerifiCode() {
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

	claOrgID := this.GetString(":cla_org_id")
	adminEmail := this.GetString(":email")

	_, emailCfg, err := getEmailConfig(claOrgID)
	if err != nil {
		reason = err
		return
	}

	ec, err := email.GetEmailClient(emailCfg.Platform)
	if err != nil {
		reason = err
		errCode = util.ErrUnknownEmailPlatform
		statusCode = 500
		return
	}

	expiry := conf.AppConfig.VerificationCodeExpiry
	code, err := models.CreateCorporationSigningVerifCode(adminEmail, expiry)
	if err != nil {
		reason = err
		return
	}

	msg := email.EmailMessage{
		To:      []string{adminEmail},
		Content: code,
		Subject: "verification code",
	}
	if err := ec.SendEmail(emailCfg.Token, &msg); err != nil {
		reason = err
		errCode = util.ErrSendingEmail
		statusCode = 500
		return
	}

	body = map[string]int64{
		"expiry": expiry,
	}
}
