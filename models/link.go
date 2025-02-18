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

package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

type OrgInfo = dbmodels.OrgInfo
type OrgRepo = dbmodels.OrgRepo

type LinkCreateOption struct {
	Platform string `json:"platform"`
	OrgID    string `json:"org_id"`
	RepoID   string `json:"repo_id"`
	OrgAlias string `json:"org_alias"`
	OrgEmail string `json:"org_email"`

	IndividualCLA *CLACreateOpt `json:"individual_cla"`
	CorpCLA       *CLACreateOpt `json:"corp_cla"`

	orgEmailInfo *dbmodels.OrgEmailCreateInfo `json:"-"`
}

func (this *LinkCreateOption) Validate(langs map[string]bool) IModelError {
	individualcla := this.IndividualCLA
	corpCLA := this.CorpCLA

	if (individualcla == nil) && (corpCLA == nil) {
		return newModelError(
			ErrMissgingCLA,
			fmt.Errorf("must specify one of individual and corp clas"),
		)
	}

	if individualcla != nil {
		if err := individualcla.Validate("", langs); err != nil {
			return err
		}
	}

	if corpCLA != nil {
		if err := corpCLA.Validate(dbmodels.ApplyToCorporation, langs); err != nil {
			return err
		}
	}

	orgEmail, err := dbmodels.GetDB().GetOrgEmailInfo(this.OrgEmail)
	if err != nil {
		if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
			return newModelError(ErrOrgEmailNotExists, err)
		}
		return parseDBError(err)
	}
	this.orgEmailInfo = orgEmail

	return nil
}

func (this LinkCreateOption) Create(linkID, submitter string) IModelError {
	info := dbmodels.LinkCreateOption{}
	info.LinkID = linkID
	info.Platform = this.Platform
	info.OrgID = this.OrgID
	info.RepoID = this.RepoID
	info.OrgEmail = *this.orgEmailInfo
	info.Submitter = submitter

	info.OrgAlias = this.OrgAlias
	if this.OrgAlias == "" {
		info.OrgAlias = this.OrgID
	}

	cla := this.IndividualCLA
	if cla != nil {
		info.IndividualCLAs = []dbmodels.CLACreateOption{
			*cla.toCLACreateOption(),
		}
	}

	cla = this.CorpCLA
	if cla != nil {
		info.CorpCLAs = []dbmodels.CLACreateOption{
			*cla.toCLACreateOption(),
		}
	}

	_, err := dbmodels.GetDB().CreateLink(&info)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrRecordExists) {
		return newModelError(ErrLinkExists, err)
	}

	return parseDBError(err)
}

func GetLinkID(orgRepo *OrgRepo) (string, IModelError) {
	b, err := dbmodels.GetDB().GetLinkID(orgRepo)
	if err == nil {
		return b, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return "", newModelError(ErrNoLink, err)
	}
	return "", parseDBError(err)
}

func Unlink(linkID string) IModelError {
	err := dbmodels.GetDB().Unlink(linkID)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return newModelError(ErrNoLink, err)
	}
	return parseDBError(err)
}

func GetOrgOfLink(linkID string) (*OrgInfo, IModelError) {
	v, err := dbmodels.GetDB().GetOrgOfLink(linkID)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}
	return v, parseDBError(err)
}

func ListLinks(platform string, orgs []string) ([]dbmodels.LinkInfo, IModelError) {
	v, err := dbmodels.GetDB().ListLinks(&dbmodels.LinkListOption{
		Platform: platform,
		Orgs:     orgs,
	})
	return v, parseDBError(err)
}

func GetAllLinks() ([]dbmodels.LinkInfo, IModelError) {
	v, err := dbmodels.GetDB().GetAllLinks()
	return v, parseDBError(err)
}
