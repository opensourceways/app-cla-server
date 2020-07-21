package models

type OrgRepo struct {
	ID          string `json:"id,omitempty"`
	Platform    string `json:"platform" required:"true"`
	OrgID       string `json:"org_id" required:"true"`
	RepoID      string `json:"repo_id" required:"true"`
	CLAID       string `json:"cla_id" required:"true"`
	CLALanguage string `json:"cla_language" required:"true"`
	MetadataID  string `json:"metadata_id,omitempty"`
	OrgEmail    string `json:"org_email,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
	Submitter   string `json:"submitter" required:"true"`
}

func (this *OrgRepo) Create() error {
	this.Enabled = true

	v, err := db.CreateOrgRepo(*this)
	if err == nil {
		this.ID = v
	}

	return err
}

func (this OrgRepo) Delete() error {
	return db.DisableOrgRepo(this.ID)
}

type OrgRepos struct {
	Org map[string][]string
}

func (this OrgRepos) List() ([]OrgRepo, error) {
	return db.ListOrgRepo(this)
}
