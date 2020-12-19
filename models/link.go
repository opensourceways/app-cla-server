package models

import (
	"fmt"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/dbmodels"
)

type LinkCreateOption struct {
	Platform string `json:"platform"`
	OrgID    string `json:"org_id"`
	RepoID   string `json:"repo_id"`
	OrgAlias string `json:"org_alias"`
	OrgEmail string `json:"org_email"`

	IndividualCLA *CLACreateOption `json:"individual_cla"`
	CorpCLA       *CLACreateOption `json:"corp_cla"`
}

func (this *LinkCreateOption) Validate() *ModelError {
	individualcla := this.IndividualCLA
	corpCLA := this.CorpCLA

	if !(individualcla != nil || corpCLA != nil) {
		return newModelError(
			ErrNoIndividualAndCorpCLA,
			fmt.Errorf("must specify one of individual and corp clas"),
		)
	}

	if individualcla != nil {
		if err := individualcla.Validate(""); err != nil {
			return err
		}
	}

	if corpCLA != nil {
		if err := corpCLA.Validate(dbmodels.ApplyToCorporation); err != nil {
			return err
		}
	}

	if _, err := dbmodels.GetDB().GetOrgEmailInfo(this.OrgEmail); err != nil {
		if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
			return newModelError(ErrOrgEmailNotExist, err)
		}
		return parseDBError(err)
	}

	return nil
}

func (this LinkCreateOption) Create(linkID, submitter string) *ModelError {
	info := dbmodels.LinkCreateOption{}
	info.LinkID = linkID
	info.Platform = this.Platform
	info.OrgID = this.OrgID
	info.RepoID = this.RepoID
	info.OrgEmail = this.OrgEmail
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

	beego.Info("dbmodels.GetDB().CreateLink")
	_, err := dbmodels.GetDB().CreateLink(&info)
	if err == nil {
		return nil
	}

	if err.IsErrorOf(dbmodels.ErrRecordExists) {
		return newModelError(ErrLinkExists, err)
	}

	return parseDBError(err)

}

func Unlink(linkID string) error {
	return dbmodels.GetDB().Unlink(linkID)
}

func ListLinks(platform string, orgs []string) ([]dbmodels.LinkInfo, *ModelError) {
	v, err := dbmodels.GetDB().ListLinks(&dbmodels.LinkListOption{
		Platform: platform,
		Orgs:     orgs,
	})
	return v, parseDBError(err)
}

func HasLink(orgRepo *dbmodels.OrgRepo) (bool, *ModelError) {
	b, err := dbmodels.GetDB().HasLink(orgRepo)
	if err != nil && err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return false, nil
	}
	return b, parseDBError(err)
}

func GetOrgOfLink(linkID string) (*dbmodels.OrgInfo, *ModelError) {
	v, err := dbmodels.GetDB().GetOrgOfLink(linkID)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return v, newModelError(ErrNoLink, err)
	}
	return v, parseDBError(err)
}
