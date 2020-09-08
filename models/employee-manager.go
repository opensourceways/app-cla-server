package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type EmployeeManagerCreateOption struct {
	CLAOrgID string   `json:"cla_org_id"`
	Emails   []string `json:"emails"`
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

func (this *EmployeeManagerCreateOption) Create() error {
	pw := "123456"

	opt := make([]dbmodels.CorporationManagerCreateOption, 0, len(this.Emails))

	for _, item := range this.Emails {
		opt = append(opt, dbmodels.CorporationManagerCreateOption{
			Role:          dbmodels.RoleManager,
			Email:         item,
			Password:      pw,
			CorporationID: util.EmailSuffixToKey(this.Emails[0]),
		})
	}

	return dbmodels.GetDB().AddCorporationManager(this.CLAOrgID, opt, 5)
}

func (this *EmployeeManagerCreateOption) Delete() error {
	opt := make([]dbmodels.CorporationManagerCreateOption, 0, len(this.Emails))

	for _, item := range this.Emails {
		opt = append(opt, dbmodels.CorporationManagerCreateOption{
			Role:  dbmodels.RoleManager,
			Email: item,
		})
	}

	return dbmodels.GetDB().DeleteCorporationManager(this.CLAOrgID, opt)
}
