package cla

import (
	"github.com/huaweicloud/golangsdk"
	"github.com/opensourceways/community-robot-lib/config"
)

type Configuration struct {
	CLA []CLARepoConfig `json:"cla,omitempty"`
}

func (c *Configuration) Validate() error {
	for _, item := range c.CLA {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Configuration) SetDefault() {
}

func (c *Configuration) CLAFor(org, repo string) *CLARepoConfig {
	if c == nil {
		return nil
	}

	v := make([]config.IPluginForRepo, 0, len(c.CLA))
	for i := range c.CLA {
		v = append(v, &c.CLA[i])
	}

	if i := config.FindConfig(org, repo, v); i >= 0 {
		return &c.CLA[i]
	}
	return nil
}

type CLARepoConfig struct {
	config.PluginForRepo

	// CLALabelYes is the cla label name for org/repos indicating
	// the cla has been signed
	CLALabelYes string `json:"cla_label_yes" required:"true"`

	// CLALabelNo is the cla label name for org/repos indicating
	// the cla has not been signed
	CLALabelNo string `json:"cla_label_no" required:"true"`

	// CLAID is the id which specifies an instance of CLA.
	CLAID string `json:"cla_id" required:"true"`

	// SignURL is the signing url for this Repo.
	SignURL string `json:"sign_url" required:"true"`

	// CheckByCommitter is one of ways to check CLA. There are two ways to check cla.
	// One is checking CLA by the email of committer, and Second is by the email of author.
	// Default is by email of author.
	CheckByCommitter bool `json:"check_by_committer"`
}

func (p CLARepoConfig) Validate() error {
	// TODO: how to validate
	if _, err := golangsdk.BuildRequestBody(p, ""); err != nil {
		return err
	}

	return p.PluginForRepo.Validate()
}
