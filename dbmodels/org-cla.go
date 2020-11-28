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

type OrgInfo struct {
	OrgRepo
	OrgAlias string `json:"org_alias"`
}

type OrgDetail struct {
	OrgInfo

	OrgEmail string `json:"org_email"`
}

type LinkInfo struct {
	OrgDetail

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
	OrgDetail
	CLAInfo *CLA
}

type LinkListOption struct {
	Platform string
	Orgs     []string
}
