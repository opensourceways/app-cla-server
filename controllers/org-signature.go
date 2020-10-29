package controllers

import (
	"fmt"
	"io/ioutil"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type OrgSignatureController struct {
	beego.Controller
}

func (this *OrgSignatureController) Prepare() {
	apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg}, &acForCodePlatformPayload{})
}

// @Title Post
// @Description upload org signature
// @Param	org_cla_id		path 	string	true		"org cla id"
// @router /:org_cla_id [post]
func (this *OrgSignatureController) Post() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "upload org signature")
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
		reason = fmt.Errorf("no need upload org signature for individual signing")
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	f, _, err := this.GetFile("signature_page")
	if err != nil {
		reason = err
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

	err = models.UploadOrgSignature(orgCLAID, data)
	if err != nil {
		reason = err
		return
	}

	body = "upload pdf of signature page successfully"
}

// @Title Get
// @Description download org signature
// @Param	org_cla_id		path 	string	true		"org cla id"
// @router /:org_cla_id [get]
func (this *OrgSignatureController) Get() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "download org signature")
	}()

	orgCLAID, err := fetchStringParameter(&this.Controller, ":org_cla_id")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	_, statusCode, errCode, reason = canAccessOrgCLA(&this.Controller, orgCLAID)
	if reason != nil {
		return
	}

	pdf, err := models.DownloadOrgSignature(orgCLAID)
	if err != nil {
		reason = err
		return
	}

	body = map[string]interface{}{
		"pdf": pdf,
	}
}

// @Title BlankSignature
// @Description get blank pdf of org signature
// @Param	language		path 	string	true		"The language which the signature applies to"
// @router /blank/:language [get]
func (this *OrgSignatureController) BlankSignature() {
	var statusCode = 0
	var errCode = ""
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, errCode, reason, body, "download blank pdf of org signature")
	}()

	language, err := fetchStringParameter(&this.Controller, ":language")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	pdf, err := models.DownloadBlankSignature(language)
	if err != nil {
		reason = err
		return
	}

	body = map[string]interface{}{
		"pdf": pdf,
	}
}
