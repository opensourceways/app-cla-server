package controllers

import (
	"fmt"
	"io/ioutil"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/util"
)

type CorporationPDFController struct {
	beego.Controller
}

func (this *CorporationPDFController) Prepare() {
	if getRouterPattern(&this.Controller) == "/v1/corporation-pdf" {
		// admin reviews pdf
		apiPrepare(&this.Controller, []string{PermissionCorporAdmin})
	} else {
		apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg})
	}
}

// @Title Upload
// @Description upload pdf of corporation signing
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 204 {int} map
// @router /:org_cla_id/:email [patch]
func (this *CorporationPDFController) Upload() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "upload corp's signing pdf")
	}()

	if err := checkAPIStringParameter(&this.Controller, []string{":org_cla_id", ":email"}); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}
	orgCLAID := this.GetString(":org_cla_id")
	corpEmail := this.GetString(":email")

	_, statusCode, errCode, reason = canAccessOrgCLA(&this.Controller, orgCLAID)
	if reason != nil {
		return
	}

	_, err := models.GetCorporationSigningBasicInfo(orgCLAID, corpEmail)
	if err != nil {
		reason = err
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
		return
	}

	err = models.UploadCorporationSigningPDF(orgCLAID, corpEmail, &data)
	if err != nil {
		reason = err
		return
	}

	body = "upload pdf of signature page successfully"
}

// @Title Download
// @Description download pdf of corporation signing
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 200 {int} map
// @router /:org_cla_id/:email [get]
func (this *CorporationPDFController) Download() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "download corp's signing pdf")
	}()

	if err := checkAPIStringParameter(&this.Controller, []string{":org_cla_id", ":email"}); err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}
	orgCLAID := this.GetString(":org_cla_id")

	_, statusCode, errCode, reason = canAccessOrgCLA(&this.Controller, orgCLAID)
	if reason != nil {
		return
	}

	pdf, err := models.DownloadCorporationSigningPDF(orgCLAID, this.GetString(":email"))
	if err != nil {
		reason = err
		return
	}

	body = map[string]interface{}{
		"pdf": pdf,
	}
}

// @Title Review
// @Description corp administrator review pdf of corporation signing
// @Success 200 {int} map
// @router / [get]
func (this *CorporationPDFController) Review() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "download corp's signing pdf")
	}()

	var ac *acForCorpManagerPayload
	ac, errCode, reason = getACOfCorpManager(&this.Controller)
	if reason != nil {
		statusCode = 401
		return
	}

	pdf, err := models.DownloadCorporationSigningPDF(ac.OrgCLAID, ac.Email)
	if err != nil {
		reason = err
		return
	}

	body = map[string]interface{}{
		"pdf": pdf,
	}
}

// @Title Preview
// @Description preview the unsinged pdf of corp
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Success 200 {int} map
// @router /:org_cla_id [get]
func (this *CorporationPDFController) Preview() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "preview the unsinged pdf of corp")
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

	if isNotCorpCLA(orgCLA) {
		reason = fmt.Errorf("not cla applied to corp")
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	cla := &models.CLA{ID: orgCLA.CLAID}
	if err := cla.Get(); err != nil {
		reason = err
		return
	}

	value := map[string]string{}
	for _, item := range cla.Fields {
		value[item.ID] = ""
	}

	signing := models.CorporationSigning{
		CorporationSigningBasicInfo: dbmodels.CorporationSigningBasicInfo{
			AdminEmail: "abc@blank_pdf.com",
		},
		Info: dbmodels.TypeSigningInfo(value),
	}

	pdf.GetPDFGenerator().GenPDFForCorporationSigning(orgCLA, &signing, cla)
	// TODO: not finished
}
