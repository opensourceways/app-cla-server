package models

import (
	"fmt"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
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

func (this *LinkCreateOption) Validate() (string, error) {
	individualcla := this.IndividualCLA
	corpCLA := this.CorpCLA

	if !(individualcla != nil || corpCLA != nil) {
		return util.ErrInvalidParameter, fmt.Errorf("must specify one of individual and corp clas")
	}

	if individualcla != nil {
		if ec, err := individualcla.Validate(""); err != nil {
			return ec, err
		}
	}

	if corpCLA != nil {
		if ec, err := corpCLA.Validate(dbmodels.ApplyToCorporation); err != nil {
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

func (this LinkCreateOption) Create(linkID, submitter string) (string, error) {
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
	return dbmodels.GetDB().CreateLink(&info)
}

func Unlink(linkID string) error {
	return dbmodels.GetDB().Unlink(linkID)
}

func ListLinks(platform string, orgs []string) ([]dbmodels.LinkInfo, error) {
	return dbmodels.GetDB().ListLinks(&dbmodels.LinkListOption{
		Platform: platform,
		Orgs:     orgs,
	})
}

func HasLink(orgRepo *dbmodels.OrgRepo) (bool, error) {
	b, err := dbmodels.GetDB().HasLink(orgRepo)
	if err != nil && dbmodels.IsErrOfDB(err, dbmodels.ErrNoDBRecord) {
		return false, nil
	}
	return b, err
}
