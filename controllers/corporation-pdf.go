package controllers

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/util"
)

type CorporationPDFController struct {
	baseController
}

func (this *CorporationPDFController) Prepare() {
	this.stopRunIfSignSerivceIsUnabled()

	if strings.HasSuffix(this.routerPattern(), "/") {
		// admin reviews pdf
		this.apiPrepare(PermissionCorpAdmin)
	} else {
		this.apiPrepare(PermissionOwnerOfOrg)
	}
}

func (this *CorporationPDFController) downloadCorpPDF(index *models.SigningIndex) *failedApiResult {
	dir := util.GenFilePath(config.AppConfig.PDFOutDir, "tmp")
	name := fmt.Sprintf("%s_%s_*.pdf", index.LinkId, index.SigningId)

	f, err := ioutil.TempFile(dir, name)
	if err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	pdf, merr := models.DownloadCorporationSigningPDF(index)
	if merr != nil {
		if merr.IsErrorOf(models.ErrNoLinkOrUnuploaed) {
			return newFailedApiResult(400, errUnuploaded, merr)
		}
		return parseModelError(merr)
	}

	if _, err = f.Write(*pdf); err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}

	f.Close()
	this.downloadFile(f.Name())
	return nil
}

// @Title Upload
// @Description upload pdf of corporation signing
// @Param	:link_id	path 	string		true		"link id"
// @Param	:signing_id	path 	string		true		"signing id"
// @Success 204 {int} map
// @router /:link_id/:signing_id [patch]
func (this *CorporationPDFController) Upload() {
	action := "upload corp's signing pdf"
	index := genSigningIndex(&this.Controller)

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(index.LinkId); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	// lock to avoid conflict with deleting corp signing
	unlock, fr := lockOnRepo(pl.orgInfo(index.LinkId))
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	defer unlock()

	if _, merr := models.GetCorpSigningBasicInfo(&index); merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}

	data, fr := this.readInputFile("pdf", config.AppConfig.MaxSizeOfCorpCLAPDF)
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if err := models.UploadCorporationSigningPDF(index, &data); err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("upload pdf of signature page successfully")
}

// @Title Download
// @Description download pdf of corporation signing
// @Param	:link_id	path 	string		true		"link id"
// @Param	:signing_id	path 	string		true		"signing id"
// @Success 200 {int} map
// @router /:link_id/:signing_id [get]
func (this *CorporationPDFController) Download() {
	action := "download corp's signing pdf"
	index := genSigningIndex(&this.Controller)

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(index.LinkId); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	if fr := this.downloadCorpPDF(&index); fr != nil {
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

	index := pl.signingIndex()
	if fr := this.downloadCorpPDF(&index); fr != nil {
		this.sendFailedResultAsResp(fr, action)
	}
}

// @Title Preview
// @Description preview the unsinged pdf of corp
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Success 200 {int} map
// @router /preview/:linkID/:language [get]
func (this *CorporationPDFController) Preview() {
	action := "preview blank pdf"
	linkID := this.GetString(":link_id")
	claLang := this.GetString(":language")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	claInfo, merr := models.GetCLAInfoToSign(linkID, claLang, dbmodels.ApplyToCorporation)
	if merr != nil {
		this.sendModelErrorAsResp(merr, action)
		return
	}
	if claInfo == nil {
		this.sendFailedResponse(400, errUnsupportedCLALang, fmt.Errorf("unsupport language"), action)
		return
	}

	claFile := genCLAFilePath(linkID, dbmodels.ApplyToCorporation, claLang, claInfo.CLAHash)

	value := map[string]string{}
	for _, item := range claInfo.Fields {
		value[item.ID] = ""
	}

	signing := models.CorporationSigning{
		CorporationSigningBasicInfo: dbmodels.CorporationSigningBasicInfo{
			AdminEmail: "test@preview_blank_pdf.com",
			Date:       util.Date(),
		},
		Info: dbmodels.TypeSigningInfo(value),
	}

	outFile, err := pdf.GetPDFGenerator().GenPDFForCorporationSigning(
		linkID, claFile, &signing, claInfo.Fields)
	if err != nil {
		this.sendFailedResponse(400, errSystemError, err, action)
		return
	}

	defer func() { os.Remove(outFile) }()
	this.downloadFile(outFile)
}
