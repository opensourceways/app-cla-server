package dbmodels

type OrgRepo struct {
	Platform string `json:"platform"`
	OrgID    string `json:"org_id"`
	RepoID   string `json:"repo_id"`
}

type OrgInfo struct {
	OrgRepo
	OrgAlias string `json:"org_alias"`
	OrgEmail string `json:"org_email"`
}

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
