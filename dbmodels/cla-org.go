package dbmodels

type CLAOrg struct {
	ID                   string `json:"id"`
	Platform             string `json:"platform"`
	OrgID                string `json:"org_id"`
	RepoID               string `json:"repo_id"`
	CLAID                string `json:"cla_id"`
	CLALanguage          string `json:"cla_language"`
	ApplyTo              string `json:"apply_to"`
	OrgEmail             string `json:"org_email"`
	Enabled              bool   `json:"enabled"`
	Submitter            string `json:"submitter"`
	OrgSignatureUploaded bool   `json:"org_signature_uploaded"`
}

type CLAOrgListOption struct {
	Platform string `json:"platform"`
	// it must specify one of OrgID and RepoID, but not both
	OrgID []string `json:"org_id"`
	// if RepoID is not empty, it is in the format of org/repo
	RepoID  string `json:"repo_id"`
	ApplyTo string `json:"apply_to"`
}
