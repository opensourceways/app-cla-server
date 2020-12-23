package controllers

import (
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/util"
)

type OrgSignatureController struct {
	baseController
}

func (this *OrgSignatureController) Prepare() {
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

	pl, err := this.tokenPayloadOfCodePlatform()
	if err != nil {
		this.sendFailedResponse(500, util.ErrSystemError, err, action)
		return
	}
	if r := pl.isOwnerOfLink(linkID); r != nil {
		this.sendFailedResponse(r.statusCode, r.errCode, r.reason, action)
		return
	}

	path := genOrgSignatureFilePath(linkID, claLang)
	this.download(path)
}

// @Title BlankSignature
// @Description get blank pdf of org signature
// @Param	language		path 	string	true		"The language which the signature applies to"
// @router /blank/:language [get]
func (this *OrgSignatureController) BlankSignature() {
	// action := "download blank pdf of org signature"

	//lang := this.GetString(":language")
	path := pdf.GetPDFGenerator().GetBlankSignaturePath()

	this.download(path)
}
