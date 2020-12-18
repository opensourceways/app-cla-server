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
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (this *EmployeeManagerCreateOption) Validate(adminEmail string) (string, error) {
	if len(this.Managers) == 0 {
		return util.ErrInvalidParameter, fmt.Errorf("no employee mangers to add/delete")
	}

	ids := map[string]bool{}
	em := map[string]bool{}
	suffix := util.EmailSuffix(adminEmail)

	for _, item := range this.Managers {
		if item.Email == "" {
			return util.ErrInvalidParameter, fmt.Errorf("missing email")
		}

		if item.Email == adminEmail {
			return util.ErrInvalidParameter, fmt.Errorf("can't add/delete administrator himself/herself")
		}

		if merr := checkEmailFormat(item.Email); merr != nil {
			return merr.ErrCode(), merr
		}

		es := util.EmailSuffix(item.Email)
		if es != suffix {
			return util.ErrNotSameCorp, fmt.Errorf("not same email suffix")
		}

		if _, ok := em[item.Email]; ok {
			return util.ErrInvalidParameter, fmt.Errorf("duplicate email:%s", item.Email)
		}
		em[item.Email] = true

		if item.ID != "" {
			if ec, err := checkManagerID(fmt.Sprintf("%s_%s", item.ID, es)); err != nil {
				return ec, err
			}

			if _, ok := ids[item.ID]; ok {
				return util.ErrInvalidParameter, fmt.Errorf("duplicate manager ID:%s", item.ID)
			}
			ids[item.ID] = true
		}
	}

	return "", nil
}

func (this *EmployeeManagerCreateOption) Create(orgCLAID string) ([]dbmodels.CorporationManagerCreateOption, error) {
	opt := make([]dbmodels.CorporationManagerCreateOption, 0, len(this.Managers))

	for _, item := range this.Managers {
		pw := util.RandStr(8, "alphanum")

		opt = append(opt, dbmodels.CorporationManagerCreateOption{
			ID:       item.ID,
			Name:     item.Name,
			Email:    item.Email,
			Password: pw,
			Role:     dbmodels.RoleManager,
		})
	}

	err := dbmodels.GetDB().AddCorporationManager(orgCLAID, opt, conf.AppConfig.EmployeeManagersNumber)
	if err != nil {
		return nil, err
	}

	es := util.EmailSuffix(opt[0].Email)
	for i := range opt {
		if opt[i].ID != "" {
			opt[i].ID = fmt.Sprintf("%s_%s", opt[i].ID, es)
		}
	}
	return opt, nil
}

func (this *EmployeeManagerCreateOption) Delete(orgCLAID string) ([]dbmodels.CorporationManagerCreateOption, error) {
	emails := make([]string, 0, len(this.Managers))

	for _, item := range this.Managers {
		emails = append(emails, item.Email)
	}

	return dbmodels.GetDB().DeleteCorporationManager(orgCLAID, emails)
}
