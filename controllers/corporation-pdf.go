package controllers

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/util"
)

type CorporationPDFController struct {
	baseController
}

func (this *CorporationPDFController) Prepare() {
	if this.routerPattern() == "/v1/corporation-pdf" {
		// admin reviews pdf
		this.apiPrepare(PermissionCorpAdmin)
	} else {
		this.apiPrepare(PermissionOwnerOfOrg)
	}
}

func (this *CorporationPDFController) downloadCorpPDF(linkID, corpEmail string) *failedApiResult {
	dir := util.GenFilePath(conf.AppConfig.PDFOutDir, "tmp")
	s := strings.ReplaceAll(util.EmailSuffix(corpEmail), ".", "_")
	name := fmt.Sprintf("%s_%s_*.pdf", linkID, s)

	f, err := ioutil.TempFile(dir, name)
	if err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	pdf, merr := models.DownloadCorporationSigningPDF(linkID, corpEmail)
	if merr != nil {
		if merr.IsErrorOf(models.ErrNoLinkOrUnuploaed) {
			return newFailedApiResult(400, errUnuploaded, merr)
		}
		return parseModelError(merr)
	}

	if _, err = f.Write(*pdf); err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}

	this.downloadFile(f.Name())
	return nil
}

// @Title Upload
// @Description upload pdf of corporation signing
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 204 {int} map
// @router /:link_id/:email [patch]
func (this *CorporationPDFController) Upload() {
	action := "upload corp's signing pdf"
	linkID := this.GetString(":link_id")
	corpEmail := this.GetString(":email")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	b, merr := models.IsCorpSigned(linkID, corpEmail)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}
	if !b {
		this.sendFailedResponse(400, errUnsigned, fmt.Errorf("not signed"), action)
		return
	}

	data, fr := this.readInputFile("pdf")
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if len(data) > (2 << 20) {
		this.sendFailedResponse(400, errTooBigPDF, fmt.Errorf("big pdf file"), action)
		return
	}

	if err := models.UploadCorporationSigningPDF(linkID, corpEmail, &data); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("upload pdf of signature page successfully")
}

// @Title Download
// @Description download pdf of corporation signing
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 200 {int} map
// @router /:link_id/:email [get]
func (this *CorporationPDFController) Download() {
	action := "download corp's signing pdf"
	linkID := this.GetString(":link_id")
	corpEmail := this.GetString(":email")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if r := pl.isOwnerOfLink(linkID); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, action)
		return
	}

	if fr := this.downloadCorpPDF(linkID, corpEmail); fr != nil {
		this.sendFailedResultAsResp(fr, action)
	}
}

// @Title Review
// @Description corp administrator review pdf of corporation signing
// @Success 200 {int} map
// @router / [get]
func (this *CorporationPDFController) Review() {
	action := "download corp's signing pdf"

	pl, fr := this.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if fr := this.downloadCorpPDF(pl.LinkID, pl.Email); fr != nil {
		this.sendFailedResultAsResp(fr, action)
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

	linkID, err := fetchStringParameter(&this.Controller, ":org_cla_id")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	var orgCLA *models.OrgCLA
	orgCLA, statusCode, errCode, reason = canAccessOrgCLA(&this.Controller, linkID)
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
