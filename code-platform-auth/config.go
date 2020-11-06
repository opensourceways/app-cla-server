package oauth

import "github.com/opensourceways/app-cla-server/oauth2"

type authConfigs struct {
	Login authConfig `json:"login" required:"true"`
	Sign  authConfig `json:"sign" required:"true"`
}

type authConfig struct {
	webRedirectDirConfig

	Configs []platformConfig `json:"platforms" required:"true"`
}

type platformConfig struct {
	Platform string `json:"platform" required:"true"`

	oauth2.Oauth2Config
}

type webRedirectDirConfig struct {
	WebRedirectDirOnSuccess string `json:"web_redirect_dir_on_success" required:"true"`
	WebRedirectDirOnFailure string `json:"web_redirect_dir_on_failure" required:"true"`
}
