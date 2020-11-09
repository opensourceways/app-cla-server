package email

type emailConfigs struct {
	webRedirectDirConfig

	Configs []platformConfig `json:"platforms" required:"true"`
}

type platformConfig struct {
	Platform    string `json:"platform" required:"true"`
	Credentials string `json:"credentials" required:"true"`
}

type webRedirectDirConfig struct {
	WebRedirectDirOnSuccess string `json:"web_redirect_dir_on_success" required:"true"`
	WebRedirectDirOnFailure string `json:"web_redirect_dir_on_failure" required:"true"`
}
