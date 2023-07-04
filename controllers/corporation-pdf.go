package controllers

import (
	"fmt"
	"os"
	"strings"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/models"
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

func (this *CorporationPDFController) downloadCorpPDF(csId string) *failedApiResult {
	pdf, merr := models.DownloadCorpPDF(csId)
	if merr != nil {
		if merr.IsErrorOf(models.ErrNoLinkOrUnuploaed) {
			return newFailedApiResult(400, errUnuploaded, merr)
		}

		return parseModelError(merr)
	}

	fn, err := util.WriteToTempFile(
		util.GenFilePath(config.AppConfig.PDFOutDir, "tmp"),
		fmt.Sprintf("%s_*.pdf", csId),
		pdf,
	)
	if err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}

	this.downloadFile(fn)

	if err := os.Remove(fn); err != nil {
		logs.Error("remove temp file failed, err: %s", err.Error())
	}

	return nil
}

// @Title Upload
// @Description upload pdf of corporation signing
// @Param  link_id     path  string  true  "link id"
// @Param  signing_id  path  string  true  "signing id"
// @Success 204 {int} map
// @router /:link_id/:signing_id [patch]
func (this *CorporationPDFController) Upload() {
	action := "upload corp's signing pdf"
	linkID := this.GetString(":link_id")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	data, fr := this.readInputFile("pdf", config.AppConfig.MaxSizeOfCorpCLAPDF)
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	err := models.UploadCorpPDF(this.GetString(":signing_id"), data)
	if err != nil {
		this.sendModelErrorAsResp(err, action)
		return
	}

	this.sendSuccessResp("upload pdf of signature page successfully")
}

// @Title Download
// @Description download pdf of corporation signing
// @Param  link_id     path  string  true  "link id"
// @Param  signing_id  path  string  true  "signing id"
// @Success 200 {int} map
// @router /:link_id/:signing_id [get]
func (this *CorporationPDFController) Download() {
	action := "download corp's signing pdf"
	linkID := this.GetString(":link_id")

	pl, fr := this.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}
	if fr := pl.isOwnerOfLink(linkID); fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	fr = this.downloadCorpPDF(this.GetString(":signing_id"))
	if fr != nil {
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

	if fr := this.downloadCorpPDF(pl.SigningId); fr != nil {
		this.sendFailedResultAsResp(fr, action)
	}
}
