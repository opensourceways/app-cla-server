package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/opensourceways/app-cla-server/util"
)

var AppConfig *appConfig

func InitAppConfig(path string) error {
	v := new(appConfig)
	if err := util.LoadFromYaml(path, v); err != nil {
		return err
	}

	v.setDefault()

	if err := v.validate(); err != nil {
		return err
	}

	AppConfig = v
	return nil
}

type appConfig struct {
	PythonBin                 string        `json:"python_bin" required:"true"`
	MaxSizeOfCorpCLAPDF       int           `json:"max_size_of_corp_cla_pdf"`
	MaxSizeOfCLAContent       int           `json:"max_size_of_cla_content"`
	VerificationCodeExpiry    int64         `json:"verification_code_expiry" required:"true"`
	APITokenExpiry            int64         `json:"api_token_expiry" required:"true"`
	APITokenKey               string        `json:"api_token_key" required:"true"`
	SymmetricEncryptionKey    string        `json:"symmetric_encryption_key" required:"true"`
	PDFOrgSignatureDir        string        `json:"pdf_org_signature_dir" required:"true"`
	PDFOutDir                 string        `json:"pdf_out_dir" required:"true"`
	CodePlatformConfigFile    string        `json:"code_platforms" required:"true"`
	EmailPlatformConfigFile   string        `json:"email_platforms" required:"true"`
	EmployeeManagersNumber    int           `json:"employee_managers_number" required:"true"`
	CLAPlatformURL            string        `json:"cla_platform_url" required:"true"`
	PasswordResetURL          string        `json:"password_reset_url" required:"true"`
	PasswordRetrievalURL      string        `json:"password_retrieval_url" required:"true"`
	PasswordRetrievalExpiry   int64         `json:"password_retrieval_expiry"`
	Mongodb                   MongodbConfig `json:"mongodb" required:"true"`
	RestrictedCorpEmailSuffix []string      `json:"restricted_corp_email_suffix"`
	MinLengthOfPassword       int           `json:"min_length_of_password"`
	MaxLengthOfPassword       int           `json:"max_length_of_password"`
	APIConfig                 apiConfig     `json:"api"     required:"true"`
	CLAConfig                 claConfig     `json:"-"`
	SigningConfig             signingConfig `json:"signing" required:"true"`
}

type MongodbConfig struct {
	MongodbConn                 string `json:"mongodb_conn" required:"true"`
	DBName                      string `json:"mongodb_db" required:"true"`
	LinkCollection              string `json:"link_collection" required:"true"`
	OrgEmailCollection          string `json:"org_email_collection" required:"true"`
	CLAPDFCollection            string `json:"cla_pdf_collection" required:"true"`
	CorpPDFCollection           string `json:"corp_pdf_collection" required:"true"`
	VCCollection                string `json:"verification_code_collection" required:"true"`
	CorpSigningCollection       string `json:"corp_signing_collection" required:"true"`
	IndividualSigningCollection string `json:"individual_signing_collection" required:"true"`
}

type apiConfig struct {
	LimitedAPIs                     []string `json:"limited_apis"`
	CookieDomain                    string   `json:"cookie_domain" required:"true"`
	CookieTimeout                   int      `json:"cookie_timeout"` // seconds
	WaitingTimeForVC                int      `json:"waiting_time_for_vc"`
	MaxRequestPerMinute             int      `json:"max_request_per_minute"`
	WebRedirectDirOnSuccessForEmail string   `json:"web_redirect_dir_on_success_for_email"`
	WebRedirectDirOnFailureForEmail string   `json:"web_redirect_dir_on_failure_for_email"`
}

func (cfg *apiConfig) setDefault() {
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

type claFileds map[string]bool

func (cfg claFileds) Number() int {
	return len(cfg)
}

func (cfg claFileds) Has(t string) bool {
	return cfg != nil && cfg[t]
}

func newCLAFields(items []string) claFileds {
	m := make(map[string]bool)
	for _, v := range items {
		m[v] = true
	}

	return claFileds(m)
}

type claConfig struct {
	AllowedCorpCLAFields       claFileds
	AllowedIndividualCLAFields claFileds
}

func (cfg *claConfig) setDefault() {
	cfg.AllowedCorpCLAFields = newCLAFields([]string{
		"fax",
		"title",
		"email",
		"address",
		"telephone",
		"authorized",
		"corporationName",
	})

	cfg.AllowedIndividualCLAFields = newCLAFields([]string{
		"name", "email",
	})
}

func (cfg *appConfig) setDefault() {
	cfg.APIConfig.setDefault()
	cfg.CLAConfig.setDefault()
	cfg.SigningConfig.setDefault()

	if cfg.MaxSizeOfCorpCLAPDF <= 0 {
		cfg.MaxSizeOfCorpCLAPDF = 5 << 20
	}

	if cfg.MaxSizeOfCLAContent <= 0 {
		cfg.MaxSizeOfCLAContent = 2 << 20
	}

	if cfg.MinLengthOfPassword <= 0 {
		cfg.MinLengthOfPassword = 8
	}

	if cfg.MaxLengthOfPassword <= 0 {
		cfg.MaxLengthOfPassword = 16
	}

	if cfg.PasswordRetrievalExpiry < 3600 {
		cfg.PasswordRetrievalExpiry = 3600
	}
}

func (cfg *appConfig) validate() error {
	if err := cfg.SigningConfig.validate(); err != nil {
		return err
	}

	if util.IsFileNotExist(cfg.PythonBin) {
		return fmt.Errorf("the file:%s is not exist", cfg.PythonBin)
	}

	if cfg.VerificationCodeExpiry <= 0 {
		return fmt.Errorf("the verification_code_expiry:%d should be bigger than 0", cfg.VerificationCodeExpiry)
	}

	if cfg.APITokenExpiry <= 0 {
		return fmt.Errorf("the apit_oken_expiry:%d should be bigger than 0", cfg.APITokenExpiry)
	}

	if cfg.EmployeeManagersNumber <= 0 {
		return fmt.Errorf("the employee_managers_number:%d should be bigger than 0", cfg.EmployeeManagersNumber)
	}

	if len(cfg.APITokenKey) < 20 {
		return fmt.Errorf("the length of api_token_key should be bigger than 20")
	}

	if _, err := util.NewSymmetricEncryption(cfg.SymmetricEncryptionKey, ""); err != nil {
		return fmt.Errorf("the symmetric encryption key is not valid, %s", err.Error())
	}

	if util.IsNotDir(cfg.PDFOrgSignatureDir) {
		return fmt.Errorf("the directory:%s is not exist", cfg.PDFOrgSignatureDir)
	}

	if util.IsNotDir(cfg.PDFOutDir) {
		return fmt.Errorf("the directory:%s is not exist", cfg.PDFOutDir)
	}

	if util.IsFileNotExist(cfg.CodePlatformConfigFile) {
		return fmt.Errorf("the file:%s is not exist", cfg.CodePlatformConfigFile)
	}

	if util.IsFileNotExist(cfg.EmailPlatformConfigFile) {
		return fmt.Errorf("the file:%s is not exist", cfg.EmailPlatformConfigFile)
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

func (cfg *appConfig) IsRestrictedEmailSuffix(emailSuffix string) bool {
	for _, suffix := range cfg.RestrictedCorpEmailSuffix {
		if strings.ToLower(suffix) == strings.ToLower(emailSuffix) {
			return true
		}
	}
	return false
}
