package controllers

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/util"
)

type OrgSignatureController struct {
	baseController
}

func (this *OrgSignatureController) Prepare() {
	if isSigningServiceNotStarted() {
		this.StopRun()
	}

	this.apiPrepare(PermissionOwnerOfOrg)
}

// @Title Get
// @Description download org signature
// @Param	org_cla_id		path 	string	true		"org cla id"
// @router /:link_id/:language [get]
func (this *OrgSignatureController) Get() {
	action := "download org signature"
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

	path := genOrgSignatureFilePath(linkID, claLang)
	if util.IsFileNotExist(path) {
		this.sendFailedResponse(400, errFileNotExists, fmt.Errorf(errFileNotExists), action)
		return
	}

	this.downloadFile(path)
}

// @Title BlankSignature
// @Description get blank pdf of org signature
// @Param	language		path 	string	true		"The language which the signature applies to"
// @router /blank/:language [get]
func (this *OrgSignatureController) BlankSignature() {
	lang := this.GetString(":language")

	path := pdf.GetPDFGenerator().GetBlankSignaturePath(lang)
	if util.IsFileNotExist(path) {
		this.sendFailedResponse(400, errFileNotExists, fmt.Errorf(errFileNotExists), "download blank signature")
		return
	}

	this.downloadFile(path)
}
