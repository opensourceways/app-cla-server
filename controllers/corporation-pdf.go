package controllers

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type CorporationPDFController struct {
	baseController
}

func (this *CorporationPDFController) Prepare() {
	if getRouterPattern(&this.Controller) == "/v1/corporation-pdf" {
		// admin reviews pdf
		this.apiPrepare(PermissionCorporAdmin)
	} else {
		this.apiPrepare(PermissionOwnerOfOrg)
	}
}

// @Title Upload
// @Description upload pdf of corporation signing
// @Param	:org_cla_id	path 	string					true		"org cla id"
// @Param	:email		path 	string					true		"email of corp"
// @Success 204 {int} map
// @router /:org_cla_id/:email [patch]
func (this *CorporationPDFController) Upload() {
	action := "upload corp's signing pdf"
	linkID := this.GetString(":link_id")
	corpEmail := this.GetString(":email")

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, action)
		return
	}
	if r := pl.isOwnerOfLink(linkID); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, action)
		return
	}

	// TODO repalce GetCorporationSigningBasicInfo by check method
	_, err = models.GetCorporationSigningBasicInfo(linkID, corpEmail)
	if err != nil {
		this.sendFailedResponse(0, "", err, action)
		return
	}

	data, fr := this.readInputFile("pdf")
	if fr != nil {
		this.sendFailedResultAsResp(fr, action)
		return
	}

	err = models.UploadCorporationSigningPDF(linkID, corpEmail, &data)
	if err != nil {
		this.sendFailedResponse(0, "", err, action)
		return
	}

	this.sendResponse("upload pdf of signature page successfully", 0)
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

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, action)
		return
	}
	if r := pl.isOwnerOfLink(linkID); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, action)
		return
	}

	pdf, err := models.DownloadCorporationSigningPDF(linkID, corpEmail)
	if err != nil {
		this.sendFailedResponse(0, "", err, action)
		return
	}

	this.sendResponse(map[string]interface{}{"pdf": pdf}, 0)
}

// @Title Review
// @Description corp administrator review pdf of corporation signing
// @Success 200 {int} map
// @router / [get]
func (this *CorporationPDFController) Review() {
	action := "download corp's signing pdf"

	pl, err := this.tokenPayloadOfCorpManager()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, action)
		return
	}

	pdf, err := models.DownloadCorporationSigningPDF(pl.LinkID, pl.Email)
	if err != nil {
		this.sendFailedResponse(0, "", err, action)
		return
	}

	this.sendResponse(map[string]interface{}{"pdf": pdf}, 0)
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

	_, err := fetchStringParameter(&this.Controller, ":org_cla_id")
	if err != nil {
		reason = err
		errCode = util.ErrInvalidParameter
		statusCode = 400
		return
	}

	/*
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
		TODO: not finished
	*/
}
