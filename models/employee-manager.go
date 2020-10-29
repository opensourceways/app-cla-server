package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type EmployeeManagerCreateOption struct {
	Managers []EmployeeManager `json:"managers"`
}

type EmployeeManager struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (this *EmployeeManagerCreateOption) Validate(adminEmail string) (string, error) {
	if len(this.Managers) == 0 {
		return util.ErrInvalidParameter, fmt.Errorf("no employee mangers to add/delete")
	}

	em := map[string]bool{}
	suffix := util.EmailSuffix(adminEmail)

	for _, item := range this.Managers {
		if item.Email == adminEmail {
			return util.ErrInvalidParameter, fmt.Errorf("can't add/delete administrator himself/herself")
		}

		if ec, err := checkEmailFormat(item.Email); err != nil {
			return ec, err
		}

		s := util.EmailSuffix(item.Email)
		if s != suffix {
			return util.ErrNotSameCorp, fmt.Errorf("the email suffixes are not same")
		}

		em[item.Email] = true
	}

	if len(em) != len(this.Managers) {
		return util.ErrInvalidParameter, fmt.Errorf("there are duplicate emails")
	}

	return "", nil
}

func (this *EmployeeManagerCreateOption) Create(orgCLAID string) ([]dbmodels.CorporationManagerCreateOption, error) {
	opt := make([]dbmodels.CorporationManagerCreateOption, 0, len(this.Managers))

	for _, item := range this.Managers {
		pw := util.RandStr(8, "alphanum")

		opt = append(opt, dbmodels.CorporationManagerCreateOption{
			Role:     dbmodels.RoleManager,
			Name:     item.Name,
			Email:    item.Email,
			Password: pw,
		})
	}

	return dbmodels.GetDB().AddCorporationManager(orgCLAID, opt, conf.AppConfig.EmployeeManagersNumber)
}

func (this *EmployeeManagerCreateOption) Delete(orgCLAID string) ([]dbmodels.CorporationManagerCreateOption, error) {
	emails := make([]string, 0, len(this.Managers))

	for _, item := range this.Managers {
		emails = append(emails, item.Email)
	}

	return dbmodels.GetDB().DeleteCorporationManager(orgCLAID, dbmodels.RoleManager, emails)
}
