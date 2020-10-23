package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type EmployeeManagerCreateOption struct {
	Emails []string `json:"emails"`
}

func (this *EmployeeManagerCreateOption) Validate(adminEmail string) (string, error) {
	if len(this.Emails) == 0 {
		return util.ErrInvalidParameter, fmt.Errorf("no employee mangers to add")
	}

	em := map[string]bool{}
	suffix := util.EmailSuffix(adminEmail)

	for _, item := range this.Emails {
		if item == adminEmail {
			return util.ErrInvalidParameter, fmt.Errorf("can't add/delete administrator himself/herself")
		}

		s := util.EmailSuffix(item)
		if s != suffix {
			return util.ErrNotSameCorp, fmt.Errorf("the email suffixes are not same")
		}

		em[item] = true
	}

	if len(em) != len(this.Emails) {
		return util.ErrInvalidParameter, fmt.Errorf("there are duplicate emails")
	}

	for _, email := range this.Emails {
		if ec, err := checkEmailFormat(email); err != nil {
			return ec, err
		}
	}

	return "", nil
}

func (this *EmployeeManagerCreateOption) Create(orgCLAID string) ([]dbmodels.CorporationManagerCreateOption, error) {
	opt := make([]dbmodels.CorporationManagerCreateOption, 0, len(this.Emails))

	for _, item := range this.Emails {
		pw := util.RandStr(8, "alphanum")

		opt = append(opt, dbmodels.CorporationManagerCreateOption{
			Role:     dbmodels.RoleManager,
			Email:    item,
			Password: pw,
		})
	}

	return dbmodels.GetDB().AddCorporationManager(orgCLAID, opt, conf.AppConfig.EmployeeManagersNumber)
}

func (this *EmployeeManagerCreateOption) Delete(orgCLAID string) ([]string, error) {
	opt := make([]dbmodels.CorporationManagerCreateOption, 0, len(this.Emails))

	for _, item := range this.Emails {
		opt = append(opt, dbmodels.CorporationManagerCreateOption{
			Role:  dbmodels.RoleManager,
			Email: item,
		})
	}

	return dbmodels.GetDB().DeleteCorporationManager(orgCLAID, opt)
}
