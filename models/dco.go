package models

type DCOLinkCreateOption struct {
	Platform string        `json:"platform"`
	OrgID    string        `json:"org_id"`
	RepoID   string        `json:"repo_id"`
	OrgAlias string        `json:"org_alias"`
	OrgEmail string        `json:"org_email"`
	DCO      *DCOCreateOpt `json:"dco"`
}

type DCOCreateOpt = struct {
	URL      string     `json:"url"`
	Fields   []CLAField `json:"fields"`
	Language string     `json:"language"`
}
