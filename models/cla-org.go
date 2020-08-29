package models

import (
	"time"

	"github.com/zengchen1024/cla-server/dbmodels"
)

type CLAOrg struct {
	ID          string    `json:"id"`
	Platform    string    `json:"platform"`
	OrgID       string    `json:"org_id"`
	RepoID      string    `json:"repo_id"`
	CLAID       string    `json:"cla_id"`
	CLALanguage string    `json:"cla_language"`
	ApplyTo     string    `json:"apply_to"`
	OrgEmail    string    `json:"org_email"`
	Enabled     bool      `json:"enabled"`
	Submitter   string    `json:"submitter"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (this *CLAOrg) Create() error {
	this.Enabled = true

	p := dbmodels.CLAOrg{}
	if err := copyBetweenStructs(this, &p); err != nil {
		return err
	}

	v, err := dbmodels.GetDB().CreateBindingBetweenCLAAndOrg(p)
	if err == nil {
		this.ID = v
	}

	return err
}

func (this CLAOrg) Delete() error {
	return dbmodels.GetDB().DeleteBindingBetweenCLAAndOrg(this.ID)
}

func (this *CLAOrg) Get() error {
	v, err := dbmodels.GetDB().GetBindingBetweenCLAAndOrg(this.ID)
	if err != nil {
		return err
	}
	return copyBetweenStructs(&v, this)
}

type CLAOrgListOption struct {
	Platform string `json:"platform"`
	OrgID    string `json:"org_id"`
	RepoID   string `json:"repo_id"`
	ApplyTo  string `json:"apply_to"`
}

func (this CLAOrgListOption) List() ([]dbmodels.CLAOrg, error) {
	p := dbmodels.CLAOrgListOption{}
	if err := copyBetweenStructs(&this, &p); err != nil {
		return nil, err
	}
	return dbmodels.GetDB().ListBindingBetweenCLAAndOrg(p)
}
