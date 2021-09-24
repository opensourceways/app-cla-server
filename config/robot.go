package config

import (
	"github.com/opensourceways/app-cla-server/util"
)

func LoadRobotServiceeConfig(path string) (RobotServiceConfig, error) {
	cfg := RobotServiceConfig{}
	err := util.LoadFromYaml(path, &cfg)
	return cfg, err
}

type RobotServiceConfig struct {
	CLAPlatformURL       string                `json:"cla_platform_url" required:"true"`
	Mongodb              MongodbConfig         `json:"mongodb" required:"true"`
	PlatformRobotConfigs []PlatformRobotConfig `json:"robot_configs" required:"true"`
}

type PlatformRobotConfig struct {
	// CodePlatform is the code platform name, sucha as github, gitee.
	CodePlatform string `json:"code_platform" required:"true"`

	// RobotTokenFile is the file path which describes the token of robot.
	RobotTokenFile string `json:"robot_token_file" required:"true"`

	// HMacSecretFile is the file path which describes the hmac info.
	HMacSecretFile string `json:"hmac_secret_file" required:"true"`

	// LitePRCommitter is the config for lite pr committer.
	LitePRCommitter LitePRCommiter `json:"lite_pr_committer" required:"true"`

	// FAQOfCheckingByAuthor is the url of faq which describes the details of checking CLA by author of commit.
	FAQOfCheckingByAuthor string `json:"faq_of_checking_by_author"`

	// FAQOfCheckingByAuthor is the url of faq which describes the details of checking CLA by committer of commit.
	FAQOfCheckingByCommitter string `json:"faq_of_checking_by_committer"`

	// CLARepoConfigFile is the file path which describes the info of CLA for each repos.
	CLARepoConfigFile string `json:"cla_repo_config_file" required:"true"`
}

type LitePRCommiter struct {
	// Email is the one of committer in a commit when a PR is lite
	Email string `json:"email" required:"true"`

	// Name is the one of committer in a commit when a PR is lite
	Name string `json:"name" required:"true"`
}

func (l LitePRCommiter) IsLitePR(email, name string) bool {
	return email == l.Email || name == l.Name
}

func FindPlatformRobotConfig(platform string, cfgs []PlatformRobotConfig) (int, bool) {
	for i := range cfgs {
		if cfgs[i].CodePlatform == platform {
			return i, true
		}
	}
	return 0, false
}
