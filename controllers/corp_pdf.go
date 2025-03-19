package controllers

import (
	"fmt"
	"os"
	"strings"

	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

const (
	corpCLAFileType     = "pdf"
	fileNameOfUploading = "pdf"
)

type CorporationPDFController struct {
	baseController
}

func (ctl *CorporationPDFController) Prepare() {
	if strings.HasSuffix(ctl.routerPattern(), "/") {
		// admin reviews pdf
		ctl.apiPrepare(PermissionCorpAdmin)
	} else {
		ctl.apiPrepare(PermissionOwnerOfOrg)
	}
}

func (ctl *CorporationPDFController) downloadCorpPDF(userId, csId string) *failedApiResult {
	pdf, merr := models.DownloadCorpPDF(userId, csId)
	if merr != nil {
		if merr.IsErrorOf(models.ErrNoLinkOrUnuploaed) {
			return newFailedApiResult(400, errUnuploaded, merr)
		}

		return parseModelError(merr)
	}

	fn, err := util.WriteToTempFile(
		config.PDFDownloadDir,
		fmt.Sprintf("%s_*.pdf", csId),
		pdf,
	)
	if err != nil {
		return newFailedApiResult(500, errSystemError, err)
	}

	ctl.downloadFile(fn)

	if err := os.Remove(fn); err != nil {
		logs.Error("remove temp file failed, err: %s", err.Error())
	}

	return nil
}

// @Title Upload
// @Description upload pdf of corporation signing
// @Tags CorpPDF
// @Accept json
// @Param  link_id     path  string  true  "link id"
// @Param  signing_id  path  string  true  "signing id"
// @Success 201 {object} controllers.respData
// @router /:link_id/:signing_id [post]
func (ctl *CorporationPDFController) Upload() {
	signingId := ctl.GetString(":signing_id")

	action := "community manager uploads pdf of corp CLA sign: " + signingId

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	data, fr := ctl.readInputFile(
		fileNameOfUploading, config.MaxSizeOfCorpCLAPDF, corpCLAFileType,
	)
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	err := models.UploadCorpPDF(pl.UserId, signingId, data)
	if err != nil {
		ctl.sendModelErrorAsResp(err, action)
	} else {
		ctl.sendSuccessResp(action, "successfully")
	}
}

// @Title Download
// @Description download pdf of corporation signing
// @Tags CorpPDF
// @Accept json
// @Param  link_id     path  string  true  "link id"
// @Param  signing_id  path  string  true  "signing id"
// @Success 200
// @router /:link_id/:signing_id [get]
func (ctl *CorporationPDFController) Download() {
	signingId := ctl.GetString(":signing_id")
	action := "community manager downloads pdf of corp CLA sign: " + signingId

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	fr = ctl.downloadCorpPDF(pl.UserId, ctl.GetString(":signing_id"))
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
	}
}

// @Title Review
// @Description corp administrator review pdf of corporation signing
// @Tags CorpPDF
// @Accept json
// @Success 200
// @router / [get]
func (ctl *CorporationPDFController) Review() {
	action := "corp admin downloads pdf"

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if fr := ctl.downloadCorpPDF(pl.UserId, pl.SigningId); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
	}
}
