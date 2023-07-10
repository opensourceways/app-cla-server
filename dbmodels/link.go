package dbmodels

import "fmt"

type LinkCreateOption struct {
	LinkID    string `json:"link_id"`
	Submitter string `json:"submitter"`

	OrgRepo
	OrgAlias string `json:"org_alias"`

	OrgEmail OrgEmailCreateInfo `json:"org_email"`

	IndividualCLAs []CLACreateOption `json:"individual_clas"`
	CorpCLAs       []CLACreateOption `json:"corp_clas"`
}

type LinkListOption struct {
	Platform string
	Orgs     []string
}

type LinkInfo struct {
	OrgInfo

	LinkID    string `json:"link_id"`
	Submitter string `json:"submitter"`
}

type CLAOfLink struct {
	IndividualCLAs []CLADetail `json:"individual_clas"`
	CorpCLAs       []CLADetail `json:"corp_clas"`
}

type OrgRepo struct {
	Platform string `json:"platform"`
	OrgID    string `json:"org_id"`
}

func (p *OrgRepo) ProjectURL() string {
	return fmt.Sprintf("https://%s.com/%s", p.Platform, p.OrgID)
}

type OrgInfo struct {
	OrgRepo
	OrgAlias         string `json:"org_alias"`
	OrgEmail         string `json:"org_email"`
	OrgEmailPlatform string `json:"org_email_platform"`
}
