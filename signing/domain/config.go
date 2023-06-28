package domain

var config Config

func Init(cfg *Config) {
	config = *cfg
}

type Config struct {
	// VerificationCodeExpiry is the one in seconds
	VerificationCodeExpiry       int64 `json:"verification_code_expiry"`
	MaxNumOfEmployeeManager      int   `json:"max_num_of_employee_manager"`
	MinNumOfSameEmailDomainParts int   `json:"min_num_of_same_email_domain_parts"`
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
}
