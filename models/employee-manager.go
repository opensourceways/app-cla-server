package models

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/config"
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

func (this *EmployeeManagerCreateOption) ValidateWhenAdding(linkID, adminEmail string) IModelError {
	if len(this.Managers) == 0 {
		return newModelError(ErrEmptyPayload, fmt.Errorf("no employee mangers"))
	}

	managers, err := ListCorporationManagers(linkID, adminEmail, dbmodels.RoleManager)
	if err != nil {
		return err
	}

	if len(this.Managers)+len(managers) > config.AppConfig.EmployeeManagersNumber {
		return newModelError(ErrManyEmployeeManagers, fmt.Errorf("too many employee managers"))
	}

	ids := map[string]bool{}
	em := map[string]bool{}
	for i := range managers {
		item := &managers[i]
		ids[item.ID] = true
		em[item.Email] = true
	}

	domains, err := ListCorpEmailDomain(linkID, adminEmail)
	if err != nil {
		return err
	}
	if len(domains) == 0 {
		return newModelError(ErrSystemError, fmt.Errorf("no domains"))
	}
	m := map[string]bool{}
	for _, i := range domains {
		m[i] = true
	}

	for i := range this.Managers {
		item := &this.Managers[i]

		if err := checkEmailFormat(item.Email); err != nil {
			return err
		}

		if !m[util.EmailSuffix(item.Email)] {
			return newModelError(ErrNotSameCorp, fmt.Errorf("not same email domain"))
		}

		if item.Email == adminEmail {
			return newModelError(ErrAdminAsManager, fmt.Errorf("can't add administrator"))
		}

		if _, ok := em[item.Email]; ok {
			return newModelError(ErrCorpManagerExists, fmt.Errorf("duplicate email:%s", item.Email))
		}
		em[item.Email] = true

		if err := checkManagerID(fmt.Sprintf("%s_%s", item.ID, domains[0])); err != nil {
			return err
		}

		if _, ok := ids[item.ID]; ok {
			return newModelError(ErrDuplicateManagerID, fmt.Errorf("duplicate manager ID:%s", item.ID))
		}
		ids[item.ID] = true
	}

	return nil
}

func (this *EmployeeManagerCreateOption) Create(linkID string) ([]dbmodels.CorporationManagerCreateOption, IModelError) {
	opt := make([]dbmodels.CorporationManagerCreateOption, 0, len(this.Managers))

	for i := range this.Managers {
		item := &this.Managers[i]
		pw := util.RandStr(8, "alphanum")

		opt = append(opt, dbmodels.CorporationManagerCreateOption{
			ID:       item.ID,
			Name:     item.Name,
			Email:    item.Email,
			Password: pw,
			Role:     dbmodels.RoleManager,
		})
	}

	err := dbmodels.GetDB().AddEmployeeManager(linkID, opt)
	if err != nil {
		if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
			return nil, newModelError(ErrNoLink, err)
		}
		return nil, parseDBError(err)
	}

	es := util.EmailSuffix(opt[0].Email)
	for i := range opt {
		if opt[i].ID != "" {
			opt[i].ID = fmt.Sprintf("%s_%s", opt[i].ID, es)
		}
	}
	return opt, nil
}

func (this *EmployeeManagerCreateOption) ValidateWhenDeleting(adminEmail string) IModelError {
	if len(this.Managers) == 0 {
		return newModelError(ErrEmptyPayload, fmt.Errorf("no employee mangers"))
	}

	suffix := util.EmailSuffix(adminEmail)

	for i := range this.Managers {
		item := &this.Managers[i]

		if err := checkEmailFormat(item.Email); err != nil {
			return err
		}

		if util.EmailSuffix(item.Email) != suffix {
			return newModelError(ErrNotSameCorp, fmt.Errorf("not same email suffix"))
		}

		if item.Email == adminEmail {
			return newModelError(ErrAdminAsManager, fmt.Errorf("can't delete administrator"))
		}
	}

	return nil
}

func (this *EmployeeManagerCreateOption) Delete(linkID string) ([]dbmodels.CorporationManagerCreateOption, IModelError) {
	emails := make([]string, 0, len(this.Managers))
	es := map[string]bool{}
	for i := range this.Managers {
		email := this.Managers[i].Email
		if !es[email] {
			es[email] = true
			emails = append(emails, email)
		}
	}

	v, err := dbmodels.GetDB().DeleteEmployeeManager(linkID, emails)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return nil, newModelError(ErrNoLink, err)
	}

	return nil, parseDBError(err)
}
