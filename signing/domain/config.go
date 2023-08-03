package domain

import (
	"strings"
	"time"
)

var config Config

func Init(cfg *Config) {
	config = *cfg
}

type Config struct {
	SourceOfCLAPDF       []string `json:"source_of_cla_pdf"`
	MaxSizeOfCLAContent  int      `json:"max_size_of_cla_content"`
	FileTypeOfCLAContent string   `json:"file_type_of_cla_content"`

	// AccessTokenExpiry is the one in seconds
	AccessTokenExpiry int64 `json:"access_token_expiry"`

	// VerificationCodeExpiry is the one in seconds
	VerificationCodeExpiry       int64  `json:"verification_code_expiry"`
	InvalidCorpEmailDomain       string `json:"invalid_corp_email_domain"`
	MaxNumOfEmployeeManager      int    `json:"max_num_of_employee_manager"`
	MinNumOfSameEmailDomainParts int    `json:"min_num_of_same_email_domain_parts"`

	MaxNumOfFailedLogin int `json:"max_num_of_failed_login"`

	// interval of creating verification code. seconds.
	IntervalOfCreatingVC int `json:"interval_of_creating_vc"`
}

func (cfg *Config) InvalidCorpEmailDomains() []string {
	return strings.Split(cfg.InvalidCorpEmailDomain, ",")
}

func (cfg *Config) SetDefault() {
	if cfg.VerificationCodeExpiry <= 0 {
		cfg.VerificationCodeExpiry = 300
	}

	if cfg.MaxNumOfEmployeeManager <= 0 {
		cfg.MaxNumOfEmployeeManager = 5
	}

	if cfg.MinNumOfSameEmailDomainParts <= 0 {
		cfg.MinNumOfSameEmailDomainParts = 2
	}

	if len(cfg.SourceOfCLAPDF) == 0 {
		cfg.SourceOfCLAPDF = []string{
			"https://gitee.com", "https://github.com",
		}
	}

	if cfg.MaxSizeOfCLAContent <= 0 {
		cfg.MaxSizeOfCLAContent = 2 << 20
	}

	if cfg.FileTypeOfCLAContent == "" {
		cfg.FileTypeOfCLAContent = "pdf"
	}

	if cfg.AccessTokenExpiry <= 0 {
		cfg.AccessTokenExpiry = 3600
	}

	if cfg.MaxNumOfFailedLogin <= 0 {
		cfg.MaxNumOfFailedLogin = 5
	}

	if cfg.IntervalOfCreatingVC <= 0 {
		cfg.IntervalOfCreatingVC = 60
	}
}

func (cfg *Config) GetIntervalOfCreatingVC() time.Duration {
	return time.Duration(cfg.IntervalOfCreatingVC) * time.Second
}
