package dbmodels

import (
	"fmt"
	"strings"
)

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
	Platform string `json:"platform" required:"true"`
	OrgID    string `json:"org_id" required:"true"`
	RepoID   string `json:"repo_id" required:"true"`
}

func (this OrgRepo) OrgRepoID() string {
	if this.RepoID == "" {
		return fmt.Sprintf("%s/%s", this.Platform, this.OrgID)
	}
	return fmt.Sprintf("%s/%s/%s", this.Platform, this.OrgID, this.RepoID)
}

func (this OrgRepo) ProjectURL() string {
	if this.RepoID == "" {
		return fmt.Sprintf("https://%s.com/%s", this.Platform, this.OrgID)
	}
	return fmt.Sprintf("https://%s.com/%s/%s", this.Platform, this.OrgID, this.RepoID)
}

func ParseToOrgRepo(s string) OrgRepo {
	r := OrgRepo{}

	v := strings.Split(s, "/")
	switch len(v) {
	case 2:
		r.Platform = v[0]
		r.OrgID = v[1]
		r.RepoID = ""
	case 3:
		r.Platform = v[0]
		r.OrgID = v[1]
		r.RepoID = v[2]
	default:
		r.Platform = s
		r.OrgID = ""
		r.RepoID = ""
	}
	return r
}

type OrgInfo struct {
	OrgRepo
	OrgAlias string `json:"org_alias"`
	OrgEmail string `json:"org_email"`
}
