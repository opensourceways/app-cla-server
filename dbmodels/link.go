package dbmodels

import (
	"fmt"
	"strings"
)

type OrgRepo struct {
	Platform string `json:"platform"`
	OrgID    string `json:"org_id"`
	RepoID   string `json:"repo_id"`
}

func (this OrgRepo) String() string {
	if this.RepoID == "" {
		return fmt.Sprintf("%s/%s", this.Platform, this.OrgID)
	}
	return fmt.Sprintf("%s/%s/%s", this.Platform, this.OrgID, this.RepoID)
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

type LinkInfo struct {
	OrgInfo

	LinkID    string `json:"link_id"`
	Submitter string `json:"submitter"`
}

type CLAOfLink struct {
	IndividualCLAs []CLA `json:"individual_clas"`
	CorpCLAs       []CLA `json:"corp_clas"`
}

type LinkCreateOption struct {
	LinkInfo

	CLAOfLink
}

type OrgCLAForSigning struct {
	OrgInfo
	CLAInfo *CLA
}

type LinkListOption struct {
	Platform string
	Orgs     []string
}
