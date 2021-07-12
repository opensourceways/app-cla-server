package email

import "fmt"

const (
	initAuthType    = "init_auth"
	reauthorizeType = "reauth"
)

func validAuthType() map[string]bool {
	return map[string]bool{
		initAuthType:    true,
		reauthorizeType: true,
	}
}

func IsValidAuthType(t string) bool {
	_, ok := validAuthType()[t]
	return ok
}

func IsReauthType(t string) bool {
	return t == reauthorizeType
}

type emailConfigs struct {
	WebRedirectDirConfigs map[string]webRedirectDirConfig `json:"auth_redirect_config" required:"true"`
	Configs               []platformConfig                `json:"platforms" required:"true"`
}

func (e emailConfigs) validate() error {
	t := validAuthType()
	for k, _ := range e.WebRedirectDirConfigs {
		if _, ok := t[k]; !ok {
			return fmt.Errorf("invalid auth type:%s", k)
		}
	}

	for _, pc := range e.Configs {
		if err := pc.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (e emailConfigs) redirectWebDir(success bool, t string) string {
	w, ok := e.WebRedirectDirConfigs[t]
	if !ok {
		return ""
	}
	return w.redirectDir(success)
}

type webRedirectDirConfig struct {
	OnSuccess string `json:"on_success" required:"true"`
	OnFailure string `json:"on_failure" required:"true"`
}

func (c *webRedirectDirConfig) redirectDir(success bool) string {
	if success {
		return c.OnSuccess
	}
	return c.OnFailure
}

type platformConfig struct {
	Platform     string            `json:"platform" required:"true"`
	Credentials  string            `json:"credentials" required:"true"`
	RedirectURLS map[string]string `json:"redirect_urls" required:"true"`
}

func (p platformConfig) validate() error {
	t := validAuthType()
	for k, _ := range p.RedirectURLS {
		if _, ok := t[k]; !ok {
			return fmt.Errorf("invalid auth type:%s", k)
		}
	}
	return nil
}

func (p platformConfig) redirectURL(t string) string {
	return p.RedirectURLS[t]
}
