package oauth

import (
	"io/ioutil"

	"github.com/huaweicloud/golangsdk"
	"sigs.k8s.io/yaml"

	"github.com/opensourceways/app-cla-server/oauth2"
)

type authConfigs struct {
	Login authConfig `json:"login" required:"true"`
	Sign  authConfig `json:"sign" required:"true"`
}

type authConfig struct {
	oauth2.Oauth2Config

	WebRedirectDir string `json:"web_redirect_dir,omitempty"`
}

func loadFromYaml(path string) (*authConfigs, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &authConfigs{}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, err
	}

	_, err = golangsdk.BuildRequestBody(cfg, "")
	return cfg, err
}
