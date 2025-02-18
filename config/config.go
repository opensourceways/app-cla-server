/*
 * Copyright (C) 2021. Huawei Technologies Co., Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/util"
)

var AppConfig = &appConfig{}

func InitAppConfig(path string) error {
	cfg := AppConfig
	if err := util.LoadFromYaml(path, cfg); err != nil {
		return err
	}

	cfg.setDefault()

	return cfg.validate()
}

type appConfig struct {
	PythonBin                string        `json:"python_bin" required:"true"`
	CLAFieldsNumber          int           `json:"cla_fields_number" required:"true"`
	MaxSizeOfCorpCLAPDF      int           `json:"max_size_of_corp_cla_pdf"`
	MaxSizeOfOrgSignaturePDF int           `json:"max_size_of_org_signature_pdf"`
	MinLengthOfPassword      int           `json:"min_length_of_password"`
	MaxLengthOfPassword      int           `json:"max_length_of_password"`
	VerificationCodeExpiry   int64         `json:"verification_code_expiry" required:"true"`
	APITokenExpiry           int64         `json:"api_token_expiry" required:"true"`
	APITokenKey              string        `json:"api_token_key" required:"true"`
	SymmetricEncryptionKey   string        `json:"symmetric_encryption_key" required:"true"`
	SymmetricEncryptionNonce string        `json:"symmetric_encryption_nonce" required:"true"`
	PDFOrgSignatureDir       string        `json:"pdf_org_signature_dir" required:"true"`
	PDFOutDir                string        `json:"pdf_out_dir" required:"true"`
	CodePlatformConfigFile   string        `json:"code_platforms" required:"true"`
	EmailPlatformConfigFile  string        `json:"email_platforms" required:"true"`
	EmployeeManagersNumber   int           `json:"employee_managers_number" required:"true"`
	CLAPlatformURL           string        `json:"cla_platform_url" required:"true"`
	Mongodb                  MongodbConfig `json:"mongodb" required:"true"`
	OBS                      OBS           `json:"obs" required:"true"`
}

type MongodbConfig struct {
	MongodbConn                 string `json:"mongodb_conn" required:"true"`
	DBName                      string `json:"mongodb_db" required:"true"`
	LinkCollection              string `json:"link_collection" required:"true"`
	OrgEmailCollection          string `json:"org_email_collection" required:"true"`
	CorpPDFCollection           string `json:"corp_pdf_collection" required:"true"`
	VCCollection                string `json:"verification_code_collection" required:"true"`
	CorpSigningCollection       string `json:"corp_signing_collection" required:"true"`
	IndividualSigningCollection string `json:"individual_signing_collection" required:"true"`
}

type OBS struct {
	Name           string `json:"name" required:"true"`
	Bucket         string `json:"bucket" required:"true"`
	CredentialFile string `json:"credential_file" required:"true"`
}

func (cfg *appConfig) setDefault() {
	if cfg.MaxSizeOfCorpCLAPDF <= 0 {
		cfg.MaxSizeOfCorpCLAPDF = (5 << 20)
	}
	if cfg.MaxSizeOfOrgSignaturePDF <= 0 {
		cfg.MaxSizeOfOrgSignaturePDF = (1 << 20)
	}

	if cfg.MinLengthOfPassword <= 0 {
		cfg.MinLengthOfPassword = 6
	}

	if cfg.MaxLengthOfPassword <= 0 {
		cfg.MaxLengthOfPassword = 16
	}
}

func (cfg *appConfig) validate() error {
	if util.IsFileNotExist(cfg.PythonBin) {
		return fmt.Errorf("The file:%s is not exist", cfg.PythonBin)
	}

	if cfg.CLAFieldsNumber <= 0 {
		return fmt.Errorf("The cla_fields_number:%d should be bigger than 0", cfg.CLAFieldsNumber)
	}

	if cfg.VerificationCodeExpiry <= 0 {
		return fmt.Errorf("The verification_code_expiry:%d should be bigger than 0", cfg.VerificationCodeExpiry)
	}

	if cfg.APITokenExpiry <= 0 {
		return fmt.Errorf("The apit_oken_expiry:%d should be bigger than 0", cfg.APITokenExpiry)
	}

	if cfg.EmployeeManagersNumber <= 0 {
		return fmt.Errorf("The employee_managers_number:%d should be bigger than 0", cfg.EmployeeManagersNumber)
	}

	if len(cfg.APITokenKey) < 20 {
		return fmt.Errorf("The length of api_token_key should be bigger than 20")
	}

	if _, err := util.NewSymmetricEncryption(cfg.SymmetricEncryptionKey, cfg.SymmetricEncryptionNonce); err != nil {
		return fmt.Errorf("The symmetric encryption key or nonce is invalid, %s", err.Error())
	}

	if util.IsNotDir(cfg.PDFOrgSignatureDir) {
		return fmt.Errorf("The directory:%s is not exist", cfg.PDFOrgSignatureDir)
	}

	if util.IsNotDir(cfg.PDFOutDir) {
		return fmt.Errorf("The directory:%s is not exist", cfg.PDFOutDir)
	}

	if util.IsFileNotExist(cfg.CodePlatformConfigFile) {
		return fmt.Errorf("The file:%s is not exist", cfg.CodePlatformConfigFile)
	}

	if util.IsFileNotExist(cfg.EmailPlatformConfigFile) {
		return fmt.Errorf("The file:%s is not exist", cfg.EmailPlatformConfigFile)
	}

	return nil
}
