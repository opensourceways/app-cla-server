package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type OrgCLAController struct {
	beego.Controller
}

func (this *OrgCLAController) Prepare() {
	if getRouterPattern(&this.Controller) == "/v1/org-cla/:platform/:org_id/:apply_to" {
		if getHeader(&this.Controller, headerToken) != "" {
			apiPrepare(&this.Controller, []string{PermissionIndividualSigner})
		}
	} else {
		apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg})
	}
}

// @Title Bind CLA to Org/Repo
// @Description bind cla
// @Param	body		body 	models.OrgCLA	true		"body for org-repo content"
// @Success 201 {int} models.OrgCLA
// @Failure 403 body is empty
// @router / [post]
func (this *OrgCLAController) Post() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "create org cla")
	}()

	var input models.OrgCLACreateOption
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &input); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	if ac, ec, err := getACOfCodePlatform(&this.Controller); err != nil {
		reason = err
		errCode = ec
		statusCode = 401
		return
	} else {
		input.Submitter = ac.User
	}

	if ec, err := input.Validate(); err != nil {
		reason = err
		errCode = ec
		statusCode = 400
		return
	}

	// check before creating to avoid downloading cla
	opt := models.OrgCLAListOption{
		Platform: input.Platform,
		OrgID:    input.OrgID,
		RepoID:   input.RepoID,
		ApplyTo:  input.ApplyTo,
	}
	if r, err := opt.List(); err != nil {
		reason = err
		return
	} else if len(r) > 0 {
		reason = fmt.Errorf("recreate org's cla")
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	// Create a file as a file lock for this org. The reasons are:
	// 1. A file lock needs the file exist first
	// 2. It is safe to create the file here, evet if creating a org's cla concurrently.
	//    Because it doesn't care the content of locked file
	path := util.GenFilePath(
		conf.AppConfig.PDFOrgSignatureDir,
		util.GenFileName(input.Platform, input.OrgID, input.RepoID),
	)
	if err := util.CreateLockedFile(path); err != nil {
		reason = err
		errCode = util.ErrSystemError
		statusCode = 500
		return
	}

	cla := &input.CLA
	if err := cla.DownloadCLA(); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	claID, err := cla.Create()
	if err != nil {
		reason = err
		return
	}

	uid, err := input.Create(claID)
	if err != nil {
		reason = err
		cla.Delete(claID)
		return
	}

	if input.ApplyTo == dbmodels.ApplyToIndividual {
		models.InitializeIndividualSigning(uid)
	} else {
		models.InitializeCorpSigning(uid, &models.OrgInfo{
			OrgRepo: models.OrgRepo{
				Platform: input.Platform,
				OrgID:    input.OrgID,
				RepoID:   input.RepoID,
			},
			OrgEmail: input.OrgEmail,
			OrgAlias: input.OrgAlias,
		})
	}

	body = struct {
		OrgClaID string `json:"org_cla_id"`
		models.OrgCLACreateOption
	}{
		OrgClaID:           uid,
		OrgCLACreateOption: input,
	}
}

// @Title Unbind CLA from Org/Repo
// @Description unbind cla
// @Param	uid		path 	string	true		"The uid of binding"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:org_cla_id [delete]
func (this *OrgCLAController) Delete() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body string

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "unbind cla")
	}()

	orgCLAID, err := fetchStringParameter(&this.Controller, ":org_cla_id")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	var orgCLA *models.OrgCLA
	orgCLA, statusCode, errCode, reason = canAccessOrgCLA(&this.Controller, orgCLAID)
	if reason != nil {
		return
	}

	if err := orgCLA.Delete(); err != nil {
		reason = err
		return
	}

	body = "unbinding successfully"
}

// @Title GetAll
// @Description get all org clas
// @Success 200 {object} models.OrgCLA
// @router / [get]
func (this *OrgCLAController) GetAll() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "list org cla")
	}()

	ac, ec, err := getACOfCodePlatform(&this.Controller)
	if err != nil {
		reason = err
		errCode = ec
		statusCode = 400
		return
	}

	if len(ac.Orgs) == 0 {
		reason = fmt.Errorf("not orgs")
		errCode = util.ErrSystemError
		statusCode = 500
		return
	}

	orgs := make([]string, 0, len(ac.Orgs))
	for k := range ac.Orgs {
		orgs = append(orgs, k)
	}

	r, err := models.ListOrgs(ac.Platform, orgs)
	if err != nil {
		reason = err
		return
	}

	body = r
}

// @Title GetCLA
// @Description get cla bound to org
// @Param	uid		path 	string	true		"org cla id"
// @Success 200 {object} models.CLA
// @router /:org_cla_id/cla [get]
func (this *OrgCLAController) GetCLA() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "get cla bound to org")
	}()

	orgCLAID, err := fetchStringParameter(&this.Controller, ":org_cla_id")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	var orgCLA *models.OrgCLA
	orgCLA, statusCode, errCode, reason = canAccessOrgCLA(&this.Controller, orgCLAID)
	if reason != nil {
		return
	}

	cla := &models.CLA{ID: orgCLA.CLAID}
	if err := cla.Get(); err != nil {
		reason = err
		return
	}

	body = cla
}

// @Title GetSigningPageInfo
// @Description get signing page info
// @Param	:platform	path 	string				true		"code platform"
// @Param	:org_id		path 	string				true		"org"
// @Param	repo_id		path 	string				true		"repo"
// @Param	:apply_to	path 	string				true		"apply to"
// @Success 201 {int} map
// @Failure util.ErrNoCLABindingDoc	"this org/repo has not been bound any clas"
// @Failure util.ErrNotReadyToSign	"the corp signing is not ready"
// @router /:platform/:org_id/:apply_to [get]
func (this *OrgCLAController) GetSigningPageInfo() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "fetch signing page info")
	}()

	params := []string{":platform", ":org_id", ":apply_to"}
	if err := checkAPIStringParameter(&this.Controller, params); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	org := this.GetString(":org_id")
	opt := models.OrgCLAListOption{
		Platform: this.GetString(":platform"),
		OrgID:    org,
		RepoID:   this.GetString("repo_id"),
		ApplyTo:  this.GetString(":apply_to"),
	}

	token := getHeader(&this.Controller, headerToken)
	if (token == "" && opt.ApplyTo != dbmodels.ApplyToCorporation) ||
		(token != "" && opt.ApplyTo != dbmodels.ApplyToIndividual) {
		reason = fmt.Errorf("invalid :apply_to")
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	orgCLAs, err := opt.List()
	if err != nil {
		reason = err
		return
	}
	if len(orgCLAs) == 0 {
		reason = fmt.Errorf("this org has no bound cla")
		errCode = util.ErrNoCLABindingDoc
		statusCode = 404
		return
	}

	ids := make([]string, 0, len(orgCLAs))
	m := map[string]string{}
	for _, i := range orgCLAs {
		if i.ApplyTo == dbmodels.ApplyToCorporation && !i.OrgSignatureUploaded {
			s := org
			if opt.RepoID != "" {
				s = fmt.Sprintf("%s/%s", s, opt.RepoID)
			}
			reason = fmt.Errorf("The project of '%s' is not ready to sign cla as corporation", s)
			errCode = util.ErrNotReadyToSign
			statusCode = 400
			return
		}

		ids = append(ids, i.CLAID)
		m[i.CLAID] = i.ID
	}

	clas, err := models.ListCLAByIDs(ids)
	if err != nil {
		reason = err
		return
	}

	result := map[string]interface{}{}
	for _, i := range clas {
		result[m[i.ID]] = i
	}

	body = result
}
