package models

import (
	"fmt"
	"time"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

const ActionCorporationSigning = "corporation-signing"

type CorporationSigning dbmodels.CorporationSigningInfo

type CorporationSigningCreateOption struct {
	CorporationSigning

	VerifiCode string `json:"verifi_code"`
}

func (this *CorporationSigningCreateOption) Validate() error {
	vc := dbmodels.VerificationCode{
		Email:   this.AdminEmail,
		Code:    this.VerifiCode,
		Purpose: ActionCorporationSigning,
	}

	v, err := dbmodels.GetDB().CheckVerificationCode(vc)
	if err != nil {
		return err
	}
	if !v {
		return fmt.Errorf("Verification Code is expired or wrong")
	}
	return nil
}

func (this *CorporationSigningCreateOption) Create(claOrgID string) error {
	this.Date = util.Date()

	return dbmodels.GetDB().SignAsCorporation(claOrgID, dbmodels.CorporationSigningInfo(this.CorporationSigning))
}

type CorporationSigningUdateInfo struct {
	CLAOrgID        string `json:"cla_org_id"`
	AdminEmail      string `json:"admin_email"`
	CorporationName string `json:"corporation_name"`
	Enabled         bool   `json:"enabled"`
}

func (this *CorporationSigningUdateInfo) Update() error {
	return dbmodels.GetDB().UpdateCorporationSigning(
		this.CLAOrgID, this.AdminEmail, this.CorporationName,
		dbmodels.CorporationSigningUpdateInfo{Enabled: &this.Enabled})
}

type CorporationSigningListOption dbmodels.CorporationSigningListOption

func (this CorporationSigningListOption) List() (map[string][]dbmodels.CorporationSigningDetails, error) {
	return dbmodels.GetDB().ListCorporationSigning(dbmodels.CorporationSigningListOption(this))
}

type CorporationSigningVerifCode struct {
	CLAOrgID string `json:"cla_org_id"`

	// Email is the email address of corporation
	Email string `json:"email"`
}

func (this CorporationSigningVerifCode) Create(expiry int64) (string, error) {
	code := "123456"

	vc := dbmodels.VerificationCode{
		Email:   this.Email,
		Code:    code,
		Purpose: ActionCorporationSigning,
		Expiry:  time.Now().Unix() + expiry,
	}

	err := dbmodels.GetDB().CreateVerificationCode(vc)
	return code, err
}
