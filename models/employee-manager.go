package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type EmployeeManagerCreateOption struct {
	Emails []string `json:"emails"`
}

func (this *EmployeeManagerCreateOption) Validate() error {
	if len(this.Emails) == 0 {
		return fmt.Errorf("parameter error: no user to add")
	}

	em := map[string]bool{}
	suffix := util.EmailSuffixToKey(this.Emails[0])

	for _, item := range this.Emails {
		em[item] = true

		s := util.EmailSuffixToKey(item)
		if s != suffix {
			return fmt.Errorf("parameter error: the email suffixes are not same")
		}
	}

	if len(em) != len(this.Emails) {
		return fmt.Errorf("parameter error: there are duplicate emails")
	}

	return nil
}

func (this *EmployeeManagerCreateOption) Create(claOrgID string) error {
	pw := "123456"

	opt := make([]dbmodels.CorporationManagerCreateOption, 0, len(this.Emails))

	for _, item := range this.Emails {
		opt = append(opt, dbmodels.CorporationManagerCreateOption{
			Role:     dbmodels.RoleManager,
			Email:    item,
			Password: pw,
		})
	}

	return dbmodels.GetDB().AddCorporationManager(claOrgID, opt, 5)
}

func (this *EmployeeManagerCreateOption) Delete(claOrgID string) error {
	opt := make([]dbmodels.CorporationManagerCreateOption, 0, len(this.Emails))

	for _, item := range this.Emails {
		opt = append(opt, dbmodels.CorporationManagerCreateOption{
			Role:  dbmodels.RoleManager,
			Email: item,
		})
	}

	return dbmodels.GetDB().DeleteCorporationManager(claOrgID, opt)
}
