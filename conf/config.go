package conf

import (
	"fmt"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/util"
)

var AppConfig *appConfig

type appConfig struct {
	PythonBin               string `json:"python_bin"`
	MongodbConn             string `json:"mongodb_conn"`
	DBName                  string `json:"mongodb_db"`
	CLAFieldsNumber         int    `json:"cla_fields_number"`
	VerificationCodeExpiry  int64  `json:"verification_code_expiry"`
	APITokenExpiry          int64  `json:"api_token_expiry"`
	APITokenKey             string `json:"api_token_key"`
	PDFOrgSignatureDir      string `json:"pdf_org_signature_dir"`
	PDFOutDir               string `json:"pdf_out_dir"`
	CodePlatformConfigFile  string `json:"code_platforms"`
	EmailPlatformConfigFile string `json:"email_platforms"`
	EmployeeManagersNumber  int    `json:"employee_managers_number"`
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
		MongodbConn:             beego.AppConfig.String("mongodb_conn"),
		DBName:                  beego.AppConfig.String("mongodb_db"),
		CLAFieldsNumber:         claFieldsNumber,
		VerificationCodeExpiry:  codeExpiry,
		APITokenExpiry:          tokenExpiry,
		APITokenKey:             beego.AppConfig.String("api_token_key"),
		PDFOrgSignatureDir:      beego.AppConfig.String("pdf_org_signature_dir"),
		PDFOutDir:               beego.AppConfig.String("pdf_out_dir"),
		CodePlatformConfigFile:  beego.AppConfig.String("code_platforms"),
		EmailPlatformConfigFile: beego.AppConfig.String("email_platforms"),
		EmployeeManagersNumber:  employeeMangers,
	}
	return AppConfig.validate()
}

func (this *appConfig) validate() error {
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
