package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type LinkCreateOption struct {
	Platform string `json:"platform"`
	OrgID    string `json:"org_id"`
	RepoID   string `json:"repo_id"`
	OrgAlias string `json:"org_alias"`
	OrgEmail string `json:"org_email"`

	IndividualCLA *CLACreateOption     `json:"individual_cla"`
	CorpCLA       *CorpCLACreateOption `json:"corp_cla"`
}

func (this *LinkCreateOption) Validate() (string, error) {
	if this.IndividualCLA == nil && this.CorpCLA == nil {
		return util.ErrInvalidParameter, fmt.Errorf("must specify one of individual and corp clas")
	}

	if this.IndividualCLA != nil {
		if ec, err := this.IndividualCLA.Validate(); err != nil {
			return ec, err
		}
	}

	if this.CorpCLA != nil {
		if ec, err := this.CorpCLA.Validate(); err != nil {
			return ec, err
		}
	}

	if _, err := dbmodels.GetDB().GetOrgEmailInfo(this.OrgEmail); err != nil {
		ec, err := parseErrorOfDBApi(err)
		if ec == util.ErrNoDBRecord {
			return util.ErrInvalidEmail, err
		}
		return ec, err
	}

	return "", nil
}

func (this LinkCreateOption) Create(submitter string) (string, error) {
	info := dbmodels.LinkCreateOption{}
	info.Platform = this.Platform
	info.OrgID = this.OrgID
	info.RepoID = this.RepoID
	info.OrgEmail = this.OrgEmail
	info.Submitter = submitter

	info.OrgAlias = this.OrgAlias
	if this.OrgAlias == "" {
		info.OrgAlias = this.OrgID
	}

	if this.IndividualCLA != nil {
		info.IndividualCLAs = []dbmodels.CLA{
			{
				Text:    this.IndividualCLA.content,
				CLAInfo: this.IndividualCLA.CLAInfo,
			},
		}
	}

	if this.CorpCLA != nil {
		cla := this.CorpCLA
		info.CorpCLAs = []dbmodels.CLA{
			{
				Text:         cla.content,
				CLAInfo:      cla.CLAInfo,
				OrgSignature: cla.OrgSignature,
			},
		}
	}

	return dbmodels.GetDB().CreateLink(&info)
}

func Unlink(orgRepo *dbmodels.OrgRepo) error {
	return dbmodels.GetDB().Unlink(orgRepo)
}

func ListLinks(platform string, orgs []string) ([]dbmodels.LinkInfo, error) {
	return dbmodels.GetDB().ListLinks(&dbmodels.LinkListOption{
		Platform: platform,
		Orgs:     orgs,
	})
}
