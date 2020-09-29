package controllers

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type OrgSignatureController struct {
	beego.Controller
}

func (this *OrgSignatureController) Prepare() {
	apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, nil)
}

// @Title Upload
// @Description upload pdf of signature page
// @Param	cla_org_id		path 	string	true		"the id of binding between cla and org"
// @Failure 403 body is empty
// @router /:cla_org_id [post]
func (this *OrgSignatureController) Post() {
	var statusCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	claOrgID := this.GetString(":cla_org_id")
	if claOrgID == "" {
		reason = fmt.Errorf("missing cla_org_id")
		statusCode = 400
		return
	}

	f, _, err := this.GetFile("signature_page")
	if err != nil {
		reason = err
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

	err = models.UploadOrgSignature(claOrgID, data)
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	path := util.OrgSignaturePDFFILE(
		beego.AppConfig.String("pdf_org_signature_dir"),
		claOrgID,
	)
	if !util.IsFileNotExist(path) {
		os.Remove(path)
	}
	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		reason = fmt.Errorf("Failed to write org signature pdf: %s", err.Error())
		statusCode = 500
		return
	}

	body = "upload pdf of signature page successfully"
}

// @Title Get
// @Description get org signature
// @Param	cla_org_id		path 	string	true		"The id of binding between cla and org"
// @Failure 403 :cla_org_id is empty
// @router /:cla_org_id [get]
func (this *OrgSignatureController) Get() {
	var statusCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	claOrgID := this.GetString(":cla_org_id")
	if claOrgID == "" {
		reason = fmt.Errorf("missing cla_org_id")
		statusCode = 400
		return
	}

	pdf, err := models.DownloadOrgSignature(claOrgID)
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = map[string]interface{}{
		"pdf": pdf,
	}
}

// @Title BlankSignature
// @Description get blank signature
// @Param	language		path 	string	true		"The language of blank signature"
// @Failure 403 :language is empty
// @router /blank/:language [get]
func (this *OrgSignatureController) BlankSignature() {
	var statusCode = 0
	var reason error
	var body interface{}

	defer func() {
		sendResponse1(&this.Controller, statusCode, reason, body)
	}()

	language := this.GetString(":language")
	if language == "" {
		reason = fmt.Errorf("missing language")
		statusCode = 400
		return
	}

	pdf, err := models.DownloadBlankSignature(language)
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = map[string]interface{}{
		"pdf": pdf,
	}
}
