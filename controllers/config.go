package controllers

type Config struct {
	LimitedAPIs                     []string `json:"limited_apis"`
	CookieDomain                    string   `json:"cookie_domain" required:"true"`
	CookieTimeout                   int      `json:"cookie_timeout"` // seconds
	WaitingTimeForVC                int      `json:"waiting_time_for_vc"`
	MaxRequestPerMinute             int      `json:"max_request_per_minute"`
	WebRedirectDirOnSuccessForEmail string   `json:"web_redirect_dir_on_success_for_email"`
	WebRedirectDirOnFailureForEmail string   `json:"web_redirect_dir_on_failure_for_email"`
}

func (cfg *Config) SetDefault() {
	if cfg.MaxRequestPerMinute <= 0 {
		cfg.MaxRequestPerMinute = 1
	}

	if len(cfg.LimitedAPIs) == 0 {
		cfg.LimitedAPIs = []string{
			"/v1/verification-code",
			"/v1/password-retrieval",
		}
	}

	if cfg.WaitingTimeForVC <= 0 {
		cfg.WaitingTimeForVC = 60
	}

	if cfg.CookieTimeout <= 0 {
		cfg.CookieTimeout = 1800
	}

	if cfg.WebRedirectDirOnSuccessForEmail == "" {
		cfg.WebRedirectDirOnSuccessForEmail = "/config-email"
	}

	if cfg.WebRedirectDirOnFailureForEmail == "" {
		cfg.WebRedirectDirOnFailureForEmail = "/config-email"
	}
}
