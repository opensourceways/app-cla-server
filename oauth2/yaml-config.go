package oauth2

import (
	"io/ioutil"

	"github.com/huaweicloud/golangsdk"
	"sigs.k8s.io/yaml"
)

type authConfigs struct {
	Config []authConfig `json:"config" required:"true"`
}

type authConfig struct {
	oauth2Config

	WebRedirectDir string `json:"web_redirect_dir,omitempty"`
	Purpose        string `json:"purpose,omitempty"`
}

type oauth2Config struct {
	ClientID     string   `json:"client_id" required:"true"`
	ClientSecret string   `json:"client_secret" required:"true"`
	AuthURL      string   `json:"auth_url" required:"true"`
	TokenURL     string   `json:"token_url" required:"true"`
	RedirectURL  string   `json:"redirect_url" required:"true"`
	Scope        []string `json:"scope" required:"true"`
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
