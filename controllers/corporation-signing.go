package controllers

import (
	"fmt"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
	"github.com/opensourceways/app-cla-server/worker"
)

type CorporationSigningController struct {
	beego.Controller
}

func (this *CorporationSigningController) Prepare() {
	if getRouterPattern(&this.Controller) != "/v1/corporation-signing/:org_cla_id" {
		// not signing
		apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg})
	}
}

// @Title Post
// @Description sign as corporation
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Param	body		body 	models.CorporationSigningCreateOption	true		"body for corporation signing"
// @Success 201 {int} map
// @Failure util.ErrHasSigned
// @Failure util.ErrWrongVerificationCode
// @Failure util.ErrVerificationCodeExpired
// @router /:org_cla_id [post]
func (this *CorporationSigningController) Post() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "sign as corporation")
	}()

	orgCLAID, err := fetchStringParameter(&this.Controller, ":org_cla_id")
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
	if ec, err := (&info).Validate(orgCLAID); err != nil {
		reason = err
		errCode = ec
		statusCode = 400
		return
	}

	orgCLA := &models.OrgCLA{ID: orgCLAID}
	if err := orgCLA.Get(); err != nil {
		reason = err
		return
	}
	if isNotCorpCLA(orgCLA) {
		reason = fmt.Errorf("invalid cla")
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	cla := &models.CLA{ID: orgCLA.CLAID}
	if err := cla.Get(); err != nil {
		reason = err
		return
	}

	info.Info = getSingingInfo(info.Info, cla.Fields)

	err = (&info).Create(orgCLAID, orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID)
	if err != nil {
		reason = err
		return
	}

	body = "sign successfully"

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(orgCLA, &info.CorporationSigning, cla)
}

// @Title ResendCorpSigningEmail
// @Description resend corp signing email
// @Param	:org_id		path 	string		true		"org cla id"
// @Param	:email		path 	string		true		"corp email"
// @Success 201 {int} map
// @router /:org_id/:email [post]
func (this *CorporationSigningController) ResendCorpSigningEmail() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "resend corp signing email")
	}()

	err := checkAndVerifyAPIStringParameter(&this.Controller, map[string]string{":org_id": "", ":email": ""})
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	ac, ec, err := getACOfCodePlatform(&this.Controller)
	if err != nil {
		reason = err
		errCode = ec
		statusCode = 400
		return
	}

	org, repo := parseOrgAndRepo(this.GetString(":org_id"))
	if !ac.hasOrg(org) {
		reason = fmt.Errorf("can't access org:%s", org)
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	orgCLAID, signingInfo, err := models.GetCorpSigningInfo(
		ac.Platform, org, repo, this.GetString(":email"),
	)
	if err != nil {
		reason = err
		return
	}

	orgCLA := &models.OrgCLA{ID: orgCLAID}
	if err := orgCLA.Get(); err != nil {
		reason = err
		return
	}

	cla := &models.CLA{ID: orgCLA.CLAID}
	if err := cla.Get(); err != nil {
		reason = err
		return
	}

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(
		orgCLA, (*models.CorporationSigning)(signingInfo), cla,
	)

	body = "resend email successfully"
}

// @Title GetAll
// @Description get all the corporations which have signed to a org
// @router /:org_id [get]
func (this *CorporationSigningController) GetAll() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "list corporation")
	}()

	org, err := fetchStringParameter(&this.Controller, ":org_id")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	ac, ec, err := getACOfCodePlatform(&this.Controller)
	if err != nil {
		reason = err
		errCode = ec
		statusCode = 400
		return
	}

	if !ac.hasOrg(org) {
		reason = fmt.Errorf("can't access org:%s", org)
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	opt := models.CorporationSigningListOption{
		Platform:    ac.Platform,
		OrgID:       org,
		RepoID:      this.GetString("repo_id"),
		CLALanguage: this.GetString("cla_language"),
	}

	r, err := opt.List()
	if err != nil {
		reason = err
		statusCode, errCode = convertDBError(err)
		return
	}

	corpMap := map[string]bool{}
	for k := range r {
		corps, err := models.ListCorpsWithPDFUploaded(k)
		if err != nil {
			reason = err
			return
		}
		for i := range corps {
			corpMap[corps[i]] = true
		}
	}

	type sInfo struct {
		*dbmodels.CorporationSigningDetail
		PDFUploaded bool `json:"pdf_uploaded"`
	}

	result := map[string][]sInfo{}
	for k := range r {
		items := r[k]
		details := make([]sInfo, 0, len(items))
		for i := range items {
			details = append(details, sInfo{
				CorporationSigningDetail: &items[i],
				PDFUploaded:              corpMap[util.EmailSuffix(items[i].AdminEmail)]},
			)
		}
		result[k] = details
	}
	body = result
}
