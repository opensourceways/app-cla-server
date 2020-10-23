package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"

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
			apiPrepare(&this.Controller, []string{PermissionIndividualSigner}, nil)
		}
		return
	}

	apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, &acForCodePlatformPayload{})
}

// @Title Bind CLA to Org/Repo
// @Description bind cla
// @Param	body		body 	models.CLAOrg	true		"body for org-repo content"
// @Success 201 {int} models.CLAOrg
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

	var input models.OrgRepoCreateOption
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &input); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}
	if ec, err := input.Validate(); err != nil {
		reason = err
		errCode = ec
		statusCode = 400
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
		return
	}

	body = struct {
		OrgClaID string `json:"org_cla_id"`
		models.OrgRepoCreateOption
	}{
		OrgClaID:            uid,
		OrgRepoCreateOption: input,
	}
}

// @Title Unbind CLA from Org/Repo
// @Description unbind cla
// @Param	uid		path 	string	true		"The uid of binding"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:uid [delete]
func (this *OrgCLAController) Delete() {
	var statusCode = 0
	var reason error
	var body string

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	uid := this.GetString(":uid")
	if uid == "" {
		reason = fmt.Errorf("missing binding id")
		statusCode = 400
		return
	}

	claOrg := models.CLAOrg{ID: uid}

	if err := claOrg.Delete(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "unbinding successfully"
}

// @Title GetAll
// @Description get all org clas
// @Success 200 {object} models.CLAOrg
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

	opt := models.CLAOrgListOption{
		Platform: ac.Platform,
		OrgID:    orgs,
	}

	r, err := opt.List()
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
// @router /:uid [get]
func (this *OrgCLAController) GetCLA() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "get cla bound to org")
	}()

	claOrgID, err := fetchStringParameter(&this.Controller, ":cla_org_id")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	var orgCLA *models.CLAOrg
	orgCLA, statusCode, errCode, reason = canAccessOrgCLA(&this.Controller, claOrgID)
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
	repo := this.GetString("repo_id")
	opt := models.CLAOrgListOption{
		Platform: this.GetString(":platform"),
		ApplyTo:  this.GetString(":apply_to"),
	}
	if repo != "" {
		opt.RepoID = fmt.Sprintf("%s/%s", org, repo)
	} else {
		opt.OrgID = []string{org}
	}

	token := getHeader(&this.Controller, headerToken)
	if (token == "" && opt.ApplyTo != dbmodels.ApplyToCorporation) ||
		(token != "" && opt.ApplyTo != dbmodels.ApplyToIndividual) {
		reason = fmt.Errorf("invalid :apply_to")
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	claOrgs, err := opt.List()
	if err != nil {
		reason = err
		return
	}
	if len(claOrgs) == 0 {
		reason = fmt.Errorf("this org has no bound cla")
		errCode = util.ErrNoCLABindingDoc
		statusCode = 404
		return
	}

	ids := make([]string, 0, len(claOrgs))
	m := map[string]string{}
	for _, i := range claOrgs {
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
