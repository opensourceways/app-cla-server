package email

type emailConfigs struct {
	WebRedirectDir string `json:"web_redirect_dir" required:"true"`

	Configs []platformConfig `json:"platforms" required:"true"`
}

type platformConfig struct {
	Platform    string `json:"platform" required:"true"`
	Credentials string `json:"credentials" required:"true"`
}
