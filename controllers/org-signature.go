/*
 * Copyright (C) 2021. Huawei Technologies Co., Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
