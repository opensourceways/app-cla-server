package oauth

import "github.com/opensourceways/app-cla-server/oauth2"

type authConfigs struct {
	Configs []platformConfig `json:"code_platforms" required:"true"`
}

type platformConfig struct {
	Platform string       `json:"platform" required:"true"`
	Login    actionConfig `json:"login" required:"true"`
	Sign     actionConfig `json:"sign" required:"true"`
}

type actionConfig struct {
	oauth2.Oauth2Config

	WebRedirectDir string `json:"web_redirect_dir,omitempty"`
}
