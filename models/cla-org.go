package models

import "time"

type CLAOrg struct {
	ID          string    `json:"id,omitempty"`
	Platform    string    `json:"platform" required:"true"`
	OrgID       string    `json:"org_id" required:"true"`
	RepoID      string    `json:"repo_id" required:"true"`
	CLAID       string    `json:"cla_id" required:"true"`
	CLALanguage string    `json:"cla_language" required:"true"`
	OrgEmail    string    `json:"org_email,omitempty"`
	Enabled     bool      `json:"enabled,omitempty"`
	Submitter   string    `json:"submitter" required:"true"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

func (this *CLAOrg) Create() error {
	this.Enabled = true

	v, err := db.BindCLAToOrg(*this)
	if err == nil {
		this.ID = v
	}

	return err
}

func (this CLAOrg) Delete() error {
	return db.UnbindCLAFromOrg(this.ID)
}

type CLAOrgs struct {
	Org map[string][]string
}

func (this CLAOrgs) List() ([]CLAOrg, error) {
	return db.ListBindingOfCLAAndOrg(this)
}
