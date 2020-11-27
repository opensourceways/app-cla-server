package dbmodels

const (
	ApplyToCorporation = "corporation"
	ApplyToIndividual  = "individual"
)

type OrgRepo struct {
	Platform string `json:"platform"`
	OrgID    string `json:"org_id"`
	RepoID   string `json:"repo_id"`
}

type OrgCLA struct {
	ID string `json:"id"`

	OrgCLACreateOption
}

type OrgCLACreateOption struct {
	OrgInfo
	IndividualCLAs []CLA
	CorpCLAs       []CLA
	Submitter      string `json:"submitter"`
	OrgEmail       string `json:"org_email"`
}

type OrgInfo struct {
	OrgRepo
	OrgAlias string `json:"org_alias"`
}

type OrgListOption struct {
	Platform string
	Orgs     []string
	ApplyTo  string
}
