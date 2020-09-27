package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
)

type CLAOrgController struct {
	beego.Controller
}

func (this *CLAOrgController) Prepare() {
	if getRouterPattern(&this.Controller) == "/v1/cla-org/:platform/:org_id/:apply_to" {
		apiPrepare(&this.Controller, []string{PermissionIndividualSigner}, nil)
		return
	}

	apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, &codePlatformAuth{})
}

// @Title Bind CLA to Org/Repo
// @Description bind cla
// @Param	body		body 	models.CLAOrg	true		"body for org-repo content"
// @Success 201 {int} models.CLAOrg
// @Failure 403 body is empty
// @router / [post]
func (this *CLAOrgController) Post() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	var claOrg models.CLAOrg

	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &claOrg); err != nil {
		reason = err
		statusCode = 400
		return
	}

	cla := &models.CLA{ID: claOrg.CLAID}

	if err := cla.Get(); err != nil {
		reason = fmt.Errorf("error finding the cla(id:%s), err: %v", cla.ID, err)
		statusCode = 400
		return
	}

	if cla.Language == "" {
		reason = fmt.Errorf("the language of cla(id:%s) is empty", cla.ID)
		statusCode = 500
		return
	}

	if cla.ApplyTo == "" {
		reason = fmt.Errorf("the apply_to of cla(id:%s) is empty", cla.ID)
		statusCode = 500
		return
	}

	claOrg.CLALanguage = cla.Language
	claOrg.ApplyTo = cla.ApplyTo

	if err := (&claOrg).Create(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = claOrg
}

// @Title Unbind CLA from Org/Repo
// @Description unbind cla
// @Param	uid		path 	string	true		"The uid of binding"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:uid [delete]
func (this *CLAOrgController) Delete() {
	var statusCode = 204
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
// @Description get all bindings
// @Success 200 {object} models.CLAOrg
// @router /:platform/:org_id [get]
func (this *CLAOrgController) GetAll() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	for _, p := range []string{":platform", ":org_id"} {
		if this.GetString(p) == "" {
			reason = fmt.Errorf("missing parameter of %s", p)
			statusCode = 400
			return
		}
	}
	opt := models.CLAOrgListOption{
		Platform: this.GetString(":platform"),
		OrgID:    this.GetString(":org_id"),
		RepoID:   this.GetString("repo_id"),
		ApplyTo:  this.GetString("apply_to"),
	}

	r, err := opt.List()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = r
}

// @Title GetSigningPageInfo
// @Description get signing page info
// @Success 200 {object} models.CLAOrg
// @router /:platform/:org_id/:apply_to [get]
func (this *CLAOrgController) GetSigningPageInfo() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	params := []string{":platform", ":org_id", ":apply_to"}
	if err := checkAPIStringParameter(&this.Controller, params); err != nil {
		reason = err
		statusCode = 400
		return
	}

	opt := models.CLAOrgListOption{
		Platform: this.GetString(":platform"),
		OrgID:    this.GetString(":org_id"),
		RepoID:   this.GetString("repo_id"),
		ApplyTo:  this.GetString(":apply_to"),
	}

	claOrgs, err := opt.ListForSigningPage()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}
	if len(claOrgs) == 0 {
		reason = fmt.Errorf("this org has no bound cla")
		statusCode = 404
		return
	}

	ids := make([]string, 0, len(claOrgs))
	m := map[string]string{}
	for _, i := range claOrgs {
		if i.ApplyTo == dbmodels.ApplyToCorporation && !i.OrgSignatureUploaded {
			s := opt.OrgID
			if opt.RepoID != "" {
				s = fmt.Sprintf("%s/%s", s, opt.RepoID)
			}
			reason = fmt.Errorf("The project of '%s' is not ready to sign cla as corporation", s)
			statusCode = 501
			return
		}

		ids = append(ids, i.CLAID)
		m[i.CLAID] = i.ID
	}

	clas, err := models.ListCLAByIDs(ids)
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	result := map[string]interface{}{}
	for _, i := range clas {
		result[m[i.ID]] = i
	}

	body = result
}

// @Title GetBlankPdf
// @Description get blank pdf of signature
// @router /blank-pdf/:cla_org_id [get]
func (this *CLAOrgController) GetBlankPdf() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	claOrgID := this.GetString(":cla_org_id")

	claOrg := &models.CLAOrg{ID: claOrgID}
	if err := claOrg.Get(); err != nil {
		reason = err
		statusCode = 400
		return
	}

	if claOrg.ApplyTo != dbmodels.ApplyToCorporation {
		reason = fmt.Errorf("Only can review blank pdf of corporation")
		statusCode = 400
		return
	}

	cla := &models.CLA{ID: claOrg.CLAID}
	if err := cla.Get(); err != nil {
		reason = err
		statusCode = 400
		return
	}

	value := map[string]string{}
	for _, item := range cla.Fields {
		value[item.ID] = ""
	}

	signing := models.CorporationSigning{
		CorporationSigningBasicInfo: dbmodels.CorporationSigningBasicInfo{
			AdminEmail: "abc@black_pef.com",
		},
		Info: dbmodels.TypeSigningInfo(value),
	}

	pdf.GetPDFGenerator().GenCLAPDFForCorporation(claOrg, &signing, cla)
}
