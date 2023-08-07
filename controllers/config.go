package controllers

import (
	"net/url"
	"strings"

	"github.com/opensourceways/app-cla-server/util"
)

var (
	config       Config
	orgWhitelist orgHelper
)

type orgHelper interface {
	Find(string) ([]string, error)
}

func Init(cfg *Config, h orgHelper) {
	config = *cfg
	orgWhitelist = h
}

type Config struct {
	CookieTimeout                   int    `json:"cookie_timeout"` // seconds
	MaxSizeOfCorpCLAPDF             int    `json:"max_size_of_corp_cla_pdf"`
	PasswordRetrievalExpiry         int64  `json:"password_retrieval_expiry"`
	WebRedirectDirOnSuccessForEmail string `json:"web_redirect_dir_on_success_for_email"`
	WebRedirectDirOnFailureForEmail string `json:"web_redirect_dir_on_failure_for_email"`

	CookieDomain         string `json:"cookie_domain"              required:"true"`
	PDFDownloadDir       string `json:"pdf_download_dir"           required:"true"`
	CLAPlatformURL       string `json:"cla_platform_url"           required:"true"`
	PasswordResetURL     string `json:"password_reset_url"         required:"true"`
	PasswordRetrievalURL string `json:"password_retrieval_url"     required:"true"`
}

func (cfg *Config) signingURL(linkId string) string {
	return cfg.CLAPlatformURL + linkId
}

func (cfg *Config) SetDefault() {
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
	if util.IsNotDir(cfg.PDFDownloadDir) {
		if err := util.Mkdir(cfg.PDFDownloadDir); err != nil {
			return err
		}
	}

	if _, err := url.Parse(cfg.CLAPlatformURL); err != nil {
		return err
	}

	if !strings.HasSuffix(cfg.CLAPlatformURL, "/") {
		cfg.CLAPlatformURL += "/"
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
