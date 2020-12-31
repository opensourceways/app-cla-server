package controllers

import (
	"fmt"
	"io/ioutil"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type OrgSignatureController struct {
	beego.Controller
}

func (this *OrgSignatureController) Prepare() {
	apiPrepare(&this.Controller, []string{PermissionOwnerOfOrg})
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

	if len(data) > (200 << 10) {
		reason = fmt.Errorf("big pdf file")
		errCode = util.ErrInvalidParameter
		statusCode = 400
	}

	err = models.UploadOrgSignature(orgCLAID, data)
	if err != nil {
		reason = err
		return
	}

	body = "upload pdf of signature page successfully"
}

func (this *OrgSignatureController) downloadPDF(fileName string, pdf *[]byte) *failedApiResult {
	dir := util.GenFilePath(conf.AppConfig.PDFOutDir, "tmp")
	name := fmt.Sprintf("%s_*.pdf", fileName)

	f, err := ioutil.TempFile(dir, name)
	if err != nil {
		return newFailedApiResult(500, util.ErrSystemError, err)
	}

	_, err = f.Write(*pdf)
	if err != nil {
		return newFailedApiResult(500, util.ErrSystemError, err)
	}

	downloadFile(&this.Controller, f.Name())
	return nil
}

// @Title Get
// @Description download org signature
// @Param	org_cla_id		path 	string	true		"org cla id"
// @router /:org_cla_id [get]
func (this *OrgSignatureController) Get() {
	rs := func(statusCode int, errCode string, reason error) {
		sendResponse(&this.Controller, statusCode, errCode, reason, nil, "download org signature")
	}

	orgCLAID, err := fetchStringParameter(&this.Controller, ":org_cla_id")
	if err != nil {
		rs(400, util.ErrInvalidParameter, err)
		return
	}

	_, statusCode, errCode, reason := canAccessOrgCLA(&this.Controller, orgCLAID)
	if reason != nil {
		rs(statusCode, errCode, reason)
		return
	}

	pdf, err := models.DownloadOrgSignature(orgCLAID)
	if err != nil {
		rs(0, "", err)
		return
	}

	if fr := this.downloadPDF(orgCLAID, &pdf); fr != nil {
		rs(fr.statusCode, fr.errCode, fr.reason)
	}

}

// @Title BlankSignature
// @Description get blank pdf of org signature
// @Param	language		path 	string	true		"The language which the signature applies to"
// @router /blank/:language [get]
func (this *OrgSignatureController) BlankSignature() {
	rs := func(statusCode int, errCode string, reason error) {
		sendResponse(
			&this.Controller, statusCode, errCode, reason, nil,
			"download blank pdf of org signature",
		)
	}

	language, err := fetchStringParameter(&this.Controller, ":language")
	if err != nil {
		rs(400, util.ErrInvalidParameter, err)
		return
	}

	pdf, err := models.DownloadBlankSignature(language)
	if err != nil {
		rs(0, "", err)
		return
	}

	if fr := this.downloadPDF("blank_"+language, &pdf); fr != nil {
		rs(fr.statusCode, fr.errCode, fr.reason)
	}
}
