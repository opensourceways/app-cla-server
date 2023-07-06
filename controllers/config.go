package controllers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/opensourceways/app-cla-server/util"
)

var config Config

type Config struct {
	LimitedAPIs                     []string `json:"limited_apis"`
	CookieDomain                    string   `json:"cookie_domain" required:"true"`
	CookieTimeout                   int      `json:"cookie_timeout"` // seconds
	WaitingTimeForVC                int      `json:"waiting_time_for_vc"`
	MaxRequestPerMinute             int      `json:"max_request_per_minute"`
	WebRedirectDirOnSuccessForEmail string   `json:"web_redirect_dir_on_success_for_email"`
	WebRedirectDirOnFailureForEmail string   `json:"web_redirect_dir_on_failure_for_email"`

	PDFOutDir               string `json:"pdf_out_dir" required:"true"`
	MaxSizeOfCorpCLAPDF     int    `json:"max_size_of_corp_cla_pdf"`
	CLAPlatformURL          string `json:"cla_platform_url" required:"true"`
	PasswordResetURL        string `json:"password_reset_url" required:"true"`
	PasswordRetrievalURL    string `json:"password_retrieval_url" required:"true"`
	PasswordRetrievalExpiry int64  `json:"password_retrieval_expiry"`
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

	if cfg.MaxSizeOfCorpCLAPDF <= 0 {
		cfg.MaxSizeOfCorpCLAPDF = 5 << 20
	}

	if cfg.PasswordRetrievalExpiry < 3600 {
		cfg.PasswordRetrievalExpiry = 3600
	}
}

func (cfg *Config) Validate() error {
	if util.IsNotDir(cfg.PDFOutDir) {
		return fmt.Errorf("the directory:%s is not exist", cfg.PDFOutDir)
	}

	if _, err := url.Parse(cfg.CLAPlatformURL); err != nil {
		return err
	}

	if _, err := url.Parse(cfg.PasswordRetrievalURL); err != nil {
		return err
	}

	s := cfg.PasswordResetURL
	if _, err := url.Parse(s); err != nil {
		return err
	}
	cfg.PasswordResetURL = strings.TrimSuffix(s, "/")

	return nil
}
