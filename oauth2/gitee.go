package oauth2

import (
	"fmt"

	liboauth2 "golang.org/x/oauth2"
)

func init() {
	client["gitee"] = &gitee{}
}

type gitee struct {
	cfg map[string]*liboauth2.Config
}

func (this *gitee) Initialize(credentialFile string) error {
	var c struct {
		Login oauth2Config `json:"login" required:"true"`
	}

	if err := loadFromYaml(credentialFile, &c); err != nil {
		return fmt.Errorf("Failed to load gitee oauth2 config: %s", err.Error())
	}

	this.cfg = map[string]*liboauth2.Config{
		"login": buildOauth2Config(c.Login),
	}

	return nil
}

func (this *gitee) GetToken(code, scope, target string) (*liboauth2.Token, error) {
	cfg, err := this.fetchOauth2Config(target)
	if err != nil {
		return nil, err
	}

	return fetchOauth2Token(cfg, code)
}

func (this *gitee) GetOauth2CodeURL(state, target string) (string, error) {
	cfg, err := this.fetchOauth2Config(target)
	if err != nil {
		return "", err
	}
	return getOauth2CodeURL(state, cfg), nil
}

func (this *gitee) fetchOauth2Config(target string) (*liboauth2.Config, error) {
	cfg, ok := this.cfg[target]
	if !ok {
		return nil, fmt.Errorf("Failed to fetch oauth2 config, unknown target: %s", target)
	}

	return cfg, nil
}
