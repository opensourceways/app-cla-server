package dbmodels

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
