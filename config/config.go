package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/opensourceways/app-cla-server/util"
)

var AppConfig = new(appConfig)

func InitAppConfig(path string) error {
	if err := util.LoadFromYaml(path, AppConfig); err != nil {
		return err
	}
	AppConfig.setDefault()
	return AppConfig.validate()
}

type appConfig struct {
	PythonBin                 string        `json:"python_bin" required:"true"`
	CLAFieldsNumber           int           `json:"cla_fields_number" required:"true"`
	MaxSizeOfCorpCLAPDF       int           `json:"max_size_of_corp_cla_pdf"`
	MaxSizeOfOrgSignaturePDF  int           `json:"max_size_of_org_signature_pdf"`
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
	PasswordRetrievalURL      string        `json:"password_retrieval_url" required:"true"`
	Mongodb                   MongodbConfig `json:"mongodb" required:"true"`
	RestrictedCorpEmailSuffix []string      `json:"restricted_corp_email_suffix"`
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

func (cfg *appConfig) setDefault() {
	if cfg.MaxSizeOfCorpCLAPDF <= 0 {
		cfg.MaxSizeOfCorpCLAPDF = 5 << 20
	}
	if cfg.MaxSizeOfOrgSignaturePDF <= 0 {
		cfg.MaxSizeOfOrgSignaturePDF = 1 << 20
	}
}

func (cfg *appConfig) validate() error {
	if util.IsFileNotExist(cfg.PythonBin) {
		return fmt.Errorf("the file:%s is not exist", cfg.PythonBin)
	}

	if cfg.CLAFieldsNumber <= 0 {
		return fmt.Errorf("the cla_fields_number:%d should be bigger than 0", cfg.CLAFieldsNumber)
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

	s := cfg.PasswordRetrievalURL
	if _, err := url.Parse(s); err != nil {
		return err
	}
	cfg.PasswordRetrievalURL = strings.TrimSuffix(s, "/")

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
