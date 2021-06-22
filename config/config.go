package config

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/huaweicloud/golangsdk"

	"github.com/opensourceways/app-cla-server/util"
)

var AppConfig *appConfig

type appConfig struct {
	PythonBin                 string        `json:"python_bin" required:"true"`
	CLAFieldsNumber           int           `json:"cla_fields_number" required:"true"`
	MaxSizeOfCorpCLAPDF       int           `json:"max_size_of_corp_cla_pdf"`
	MaxSizeOfOrgSignaturePDF  int           `json:"max_size_of_org_signature_pdf"`
	VerificationCodeExpiry    int64         `json:"verification_code_expiry" required:"true"`
	APITokenExpiry            int64         `json:"api_token_expiry" required:"true"`
	APITokenKey               string        `json:"api_token_key" required:"true"`
	PDFOrgSignatureDir        string        `json:"pdf_org_signature_dir" required:"true"`
	PDFOutDir                 string        `json:"pdf_out_dir" required:"true"`
	CodePlatformConfigFile    string        `json:"code_platforms" required:"true"`
	EmailPlatformConfigFile   string        `json:"email_platforms" required:"true"`
	EmployeeManagersNumber    int           `json:"employee_managers_number" required:"true"`
	CLAPlatformURL            string        `json:"cla_platform_url" required:"true"`
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

func InitAppConfig() error {
	claFieldsNumber, err := beego.AppConfig.Int("cla_fields_number")
	if err != nil {
		return err
	}

	maxSizeOfCorpCLAPDF := beego.AppConfig.DefaultInt("max_size_of_corp_cla_pdf", (2 << 20))
	maxSizeOfOrgSignaturePDF := beego.AppConfig.DefaultInt("max_size_of_org_signature_pdf", (1 << 20))

	tokenExpiry, err := beego.AppConfig.Int64("api_token_expiry")
	if err != nil {
		return err
	}

	codeExpiry, err := beego.AppConfig.Int64("verification_code_expiry")
	if err != nil {
		return err
	}

	employeeMangers, err := beego.AppConfig.Int("employee_managers_number")
	if err != nil {
		return err
	}

	AppConfig = &appConfig{
		PythonBin:                beego.AppConfig.String("python_bin"),
		CLAFieldsNumber:          claFieldsNumber,
		MaxSizeOfCorpCLAPDF:      maxSizeOfCorpCLAPDF,
		MaxSizeOfOrgSignaturePDF: maxSizeOfOrgSignaturePDF,
		VerificationCodeExpiry:   codeExpiry,
		APITokenExpiry:           tokenExpiry,
		APITokenKey:              beego.AppConfig.String("api_token_key"),
		PDFOrgSignatureDir:       beego.AppConfig.String("pdf_org_signature_dir"),
		PDFOutDir:                beego.AppConfig.String("pdf_out_dir"),
		CodePlatformConfigFile:   beego.AppConfig.String("code_platforms"),
		EmailPlatformConfigFile:  beego.AppConfig.String("email_platforms"),
		EmployeeManagersNumber:   employeeMangers,
		CLAPlatformURL:           beego.AppConfig.String("cla_platform_url"),
		Mongodb: MongodbConfig{
			MongodbConn:                 beego.AppConfig.String("mongodb::mongodb_conn"),
			DBName:                      beego.AppConfig.String("mongodb::mongodb_db"),
			LinkCollection:              beego.AppConfig.String("mongodb::link_collection"),
			OrgEmailCollection:          beego.AppConfig.String("mongodb::org_email_collection"),
			CLAPDFCollection:            beego.AppConfig.String("mongodb::cla_pdf_collection"),
			CorpPDFCollection:           beego.AppConfig.String("mongodb::corp_pdf_collection"),
			VCCollection:                beego.AppConfig.String("mongodb::verification_code_collection"),
			CorpSigningCollection:       beego.AppConfig.String("mongodb::corp_signing_collection"),
			IndividualSigningCollection: beego.AppConfig.String("mongodb::individual_signing_collection"),
		},
		RestrictedCorpEmailSuffix: beego.AppConfig.DefaultStrings("restricted_corp_email_suffix", []string{}),
	}
	return AppConfig.validate()
}

func (appConf *appConfig) validate() error {
	_, err := golangsdk.BuildRequestBody(appConf, "")
	if err != nil {
		return fmt.Errorf("config file error: %s", err.Error())
	}

	if util.IsFileNotExist(appConf.PythonBin) {
		return fmt.Errorf("The file:%s is not exist", appConf.PythonBin)
	}

	if appConf.CLAFieldsNumber <= 0 {
		return fmt.Errorf("The cla_fields_number:%d should be bigger than 0", appConf.CLAFieldsNumber)
	}

	if appConf.VerificationCodeExpiry <= 0 {
		return fmt.Errorf("The verification_code_expiry:%d should be bigger than 0", appConf.VerificationCodeExpiry)
	}

	if appConf.APITokenExpiry <= 0 {
		return fmt.Errorf("The apit_oken_expiry:%d should be bigger than 0", appConf.APITokenExpiry)
	}

	if appConf.EmployeeManagersNumber <= 0 {
		return fmt.Errorf("The employee_managers_number:%d should be bigger than 0", appConf.EmployeeManagersNumber)
	}

	if len(appConf.APITokenKey) < 20 {
		return fmt.Errorf("The length of api_token_key should be bigger than 20")
	}

	if util.IsNotDir(appConf.PDFOrgSignatureDir) {
		return fmt.Errorf("The directory:%s is not exist", appConf.PDFOrgSignatureDir)
	}

	if util.IsNotDir(appConf.PDFOutDir) {
		return fmt.Errorf("The directory:%s is not exist", appConf.PDFOutDir)

	}

	if util.IsFileNotExist(appConf.CodePlatformConfigFile) {
		return fmt.Errorf("The file:%s is not exist", appConf.CodePlatformConfigFile)
	}

	if util.IsFileNotExist(appConf.EmailPlatformConfigFile) {
		return fmt.Errorf("The file:%s is not exist", appConf.EmailPlatformConfigFile)
	}

	return nil
}

func (appConf *appConfig) IsRestrictedEmailSuffix(emailSuffix string) bool {
	for _, suffix := range appConf.RestrictedCorpEmailSuffix {
		if strings.ToLower(suffix) == strings.ToLower(emailSuffix) {
			return true
		}
	}
	return false
}
