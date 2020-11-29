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

	orgCLA := models.OrgCLA{ID: orgCLAID}
	if err := (&orgCLA).Get(); err != nil {
		reason = err
		return
	}
	if isNotCorpCLA(&orgCLA) {
		reason = fmt.Errorf("invalid cla")
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	cla := models.CLA{ID: orgCLA.CLAID}
	if err := (&cla).Get(); err != nil {
		reason = err
		return
	}

	info.Info = getSingingInfo(info.Info, cla.Fields)

	orgRepo := buildOrgRepo(orgCLA.Platform, orgCLA.OrgID, orgCLA.RepoID)
	err = (&info).Create(&orgRepo)
	if err != nil {
		reason = err
		return
	}

	body = "sign successfully"

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(orgCLA, info.CorporationSigning, cla)
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

	// TODO: resend if pdf has not been uploaded

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

	orgRepo := buildOrgRepo(ac.Platform, org, repo)
	signingInfo, err := models.GetCorporationSigningDetail(&orgRepo, this.GetString(":email"))
	if err != nil {
		reason = err
		return
	}

	orgCLA := models.OrgCLA{ID: ""}
	if err := (&orgCLA).Get(); err != nil {
		reason = err
		return
	}

	cla := models.CLA{ID: orgCLA.CLAID}
	if err := (&cla).Get(); err != nil {
		reason = err
		return
	}

	worker.GetEmailWorker().GenCLAPDFForCorporationAndSendIt(
		orgCLA,
		models.CorporationSigning{
			CorporationSigningBasicInfo: signingInfo.CorporationSigningBasicInfo,
			Info:                        signingInfo.Info,
		}, cla,
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

	orgRepo := buildOrgRepo(ac.Platform, org, this.GetString("repo_id"))
	body, reason = this.listCorpSiging(&orgRepo)
}

func (this *CorporationSigningController) listCorpSiging(orgRepo *dbmodels.OrgRepo) (interface{}, error) {
	r, err := models.ListCorporationSigning(orgRepo, this.GetString("cla_language"))
	if err != nil {
		return nil, err
	}

	managers, err := models.ListCorporationManager(orgRepo, "", dbmodels.RoleAdmin)
	if err != nil {
		return nil, err
	}

	ms := map[string]bool{}
	for _, item := range managers {
		ms[item.Email] = true
	}

	type sInfo struct {
		*dbmodels.CorporationSigningSummary
		AdminAdded bool `json:"admin_added"`
	}

	result := make([]sInfo, 0, len(r))

	for i := 0; i < len(r); i++ {
		result = append(result, sInfo{
			CorporationSigningSummary: &r[i],
			AdminAdded:                ms[r[0].AdminEmail],
		})
	}
	return result, nil
}
