package dbmodels

import (
	"fmt"
	"strings"

	"github.com/opensourceways/app-cla-server/util"
)

type OrgCLA struct {
	ID                   string `json:"id"`
	Platform             string `json:"platform"`
	OrgID                string `json:"org_id"`
	RepoID               string `json:"repo_id"`
	OrgAlias             string `json:"org_alias"`
	CLAID                string `json:"cla_id"`
	CLALanguage          string `json:"cla_language"`
	ApplyTo              string `json:"apply_to"`
	OrgEmail             string `json:"org_email"`
	Enabled              bool   `json:"enabled"`
	Submitter            string `json:"submitter"`
	OrgSignatureUploaded bool   `json:"org_signature_uploaded"`
}

type OrgCLAListOption struct {
	Platform string `json:"platform"`
	OrgID    string `json:"org_id"`
	RepoID   string `json:"repo_id"`
	ApplyTo  string `json:"apply_to"`
}

type OrgRepo struct {
	Platform string `json:"platform"`
	OrgID    string `json:"org_id"`
	RepoID   string `json:"repo_id"`
}

func (this OrgRepo) OrgRepoID() string {
	if this.RepoID == "" {
		return fmt.Sprintf("%s/%s", this.Platform, this.OrgID)
	}
	return fmt.Sprintf("%s/%s/%s", this.Platform, this.OrgID, this.RepoID)
}

func (this OrgRepo) ProjectURL() string {
	return util.ProjectURL(this.Platform, this.OrgID, this.RepoID)
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
