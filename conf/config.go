package conf

import (
	"fmt"

	"github.com/astaxie/beego"
	"github.com/huaweicloud/golangsdk"

	"github.com/opensourceways/app-cla-server/util"
)

var AppConfig *appConfig

type appConfig struct {
	PythonBin               string        `json:"python_bin" required:"true"`
	CLAFieldsNumber         int           `json:"cla_fields_number" required:"true"`
	VerificationCodeExpiry  int64         `json:"verification_code_expiry" required:"true"`
	APITokenExpiry          int64         `json:"api_token_expiry" required:"true"`
	APITokenKey             string        `json:"api_token_key" required:"true"`
	PDFOrgSignatureDir      string        `json:"pdf_org_signature_dir" required:"true"`
	PDFOutDir               string        `json:"pdf_out_dir" required:"true"`
	CodePlatformConfigFile  string        `json:"code_platforms" required:"true"`
	EmailPlatformConfigFile string        `json:"email_platforms" required:"true"`
	EmployeeManagersNumber  int           `json:"employee_managers_number" required:"true"`
	CLAPlatformURL          string        `json:"cla_platform_url" required:"true"`
	Mongodb                 MongodbConfig `json:"mongodb" required:"true"`
}

type MongodbConfig struct {
	MongodbConn string `json:"mongodb_conn" required:"true"`
	DBName      string `json:"mongodb_db" required:"true"`

	VCCollection             string `json:"verification_code_collection" required:"true"`
	OrgEmailCollection       string `json:"org_email_collection" required:"true"`
	BlankSignatureCollection string `json:"blank_signature_collection" required:"true"`

	LinkCollection              string `json:"link_collection" required:"true"`
	CorpManagerCollection       string `json:"corp_manager_collection" required:"true"`
	CorpSigningCollection       string `json:"corp_signing_collection" required:"true"`
	IndividualSigningCollection string `json:"individual_signing_collection" required:"true"`
}

func InitAppConfig() error {
	claFieldsNumber, err := beego.AppConfig.Int("cla_fields_number")
	if err != nil {
		return err
	}

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
		PythonBin:               beego.AppConfig.String("python_bin"),
		CLAFieldsNumber:         claFieldsNumber,
		VerificationCodeExpiry:  codeExpiry,
		APITokenExpiry:          tokenExpiry,
		APITokenKey:             beego.AppConfig.String("api_token_key"),
		PDFOrgSignatureDir:      beego.AppConfig.String("pdf_org_signature_dir"),
		PDFOutDir:               beego.AppConfig.String("pdf_out_dir"),
		CodePlatformConfigFile:  beego.AppConfig.String("code_platforms"),
		EmailPlatformConfigFile: beego.AppConfig.String("email_platforms"),
		EmployeeManagersNumber:  employeeMangers,
		CLAPlatformURL:          beego.AppConfig.String("cla_platform_url"),
		Mongodb: MongodbConfig{
			MongodbConn: beego.AppConfig.String("mongodb::mongodb_conn"),
			DBName:      beego.AppConfig.String("mongodb::mongodb_db"),

			VCCollection:                beego.AppConfig.String("mongodb::verification_code_collection"),
			OrgEmailCollection:          beego.AppConfig.String("mongodb::org_email_collection"),
			BlankSignatureCollection:    beego.AppConfig.String("mongodb::blank_signature_collection"),
			LinkCollection:              beego.AppConfig.String("mongodb::link_collection"),
			CorpManagerCollection:       beego.AppConfig.String("mongodb::corp_manager_collection"),
			CorpSigningCollection:       beego.AppConfig.String("mongodb::corp_signing_collection"),
			IndividualSigningCollection: beego.AppConfig.String("mongodb::individual_signing_collection"),
		},
	}
	return AppConfig.validate()
}

func (this *appConfig) validate() error {
	_, err := golangsdk.BuildRequestBody(this, "")
	if err != nil {
		return fmt.Errorf("config file error: %s", err.Error())
	}

	if util.IsFileNotExist(this.PythonBin) {
		return fmt.Errorf("The file:%s is not exist", this.PythonBin)
	}

	if this.CLAFieldsNumber <= 0 {
		return fmt.Errorf("The cla_fields_number:%d should be bigger than 0", this.CLAFieldsNumber)
	}

	if this.VerificationCodeExpiry <= 0 {
		return fmt.Errorf("The verification_code_expiry:%d should be bigger than 0", this.VerificationCodeExpiry)
	}

	if this.APITokenExpiry <= 0 {
		return fmt.Errorf("The apit_oken_expiry:%d should be bigger than 0", this.APITokenExpiry)
	}

	if this.EmployeeManagersNumber <= 0 {
		return fmt.Errorf("The employee_managers_number:%d should be bigger than 0", this.EmployeeManagersNumber)
	}

	if len(this.APITokenKey) < 20 {
		return fmt.Errorf("The length of api_token_key should be bigger than 20")
	}

	if util.IsNotDir(this.PDFOrgSignatureDir) {
		return fmt.Errorf("The directory:%s is not exist", this.PDFOrgSignatureDir)
	}

	if util.IsNotDir(this.PDFOutDir) {
		return fmt.Errorf("The directory:%s is not exist", this.PDFOutDir)

	}
	if util.IsFileNotExist(this.CodePlatformConfigFile) {
		return fmt.Errorf("The file:%s is not exist", this.CodePlatformConfigFile)
	}

	if util.IsFileNotExist(this.EmailPlatformConfigFile) {
		return fmt.Errorf("The file:%s is not exist", this.EmailPlatformConfigFile)
	}

	return nil
}
