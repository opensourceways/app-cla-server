package models

type OrgRepo struct {
	ID         string `json:"id,omitempty"`
	Platform   string `json:"platform" required:"true"`
	OrgID      string `json:"org_id" required:"true"`
	RepoID     string `json:"repo_id" required:"true"`
	CLAID      string `json:"cla_id" required:"true"`
	MetadataID string `json:"metadata_id,omitempty"`
	OrgEmail   string `json:"org_email,omitempty"`
	Enabled    bool   `json:"enabled,omitempty"`
	Submitter  string `json:"submitter" required:"true"`
}

func (this OrgRepo) Create() (OrgRepo, error) {
	this.Enabled = true

	v, err := db.CreateOrgRepo(this)
	if err == nil {
		this.ID = v
	}

	return this, err
}

func (this OrgRepo) Delete() error {
	return db.DisableOrgRepo(this.ID)
}
