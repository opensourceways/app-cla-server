package models

import "github.com/opensourceways/app-cla-server/dbmodels"

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
