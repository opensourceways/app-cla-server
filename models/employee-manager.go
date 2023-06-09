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

func (this *EmployeeManagerCreateOption) ValidateWhenAdding(
	index SigningIndex, detail *dbmodels.CorporationDetail,
) IModelError {
	if len(this.Managers) == 0 {
		return newModelError(ErrEmptyPayload, fmt.Errorf("no employee mangers"))
	}

	if len(this.Managers)+len(detail.Managers) > config.AppConfig.EmployeeManagersNumber {
		return newModelError(ErrManyEmployeeManagers, fmt.Errorf("too many employee managers"))
	}

	ids := map[string]bool{}
	em := map[string]bool{}
	for i := range detail.Managers {
		item := &detail.Managers[i]
		ids[item.ID] = true
		em[item.Email] = true
	}

	for i := range this.Managers {
		item := &this.Managers[i]

		if err := checkEmailFormat(item.Email); err != nil {
			return err
		}

		domain := util.EmailSuffix(item.Email)
		if !detail.HasDomain(domain) {
			return newModelError(ErrNotSameCorp, fmt.Errorf("not same email domain"))
		}

		if item.Email == detail.AdminEmail() {
			return newModelError(ErrAdminAsManager, fmt.Errorf("can't add administrator"))
		}

		if _, ok := em[item.Email]; ok {
			return newModelError(ErrCorpManagerExists, fmt.Errorf("duplicate email:%s", item.Email))
		}
		em[item.Email] = true

		item.ID = managerAccount(item.ID, item.Email)

		if err := checkManagerID(item.ID); err != nil {
			return err
		}

		if _, ok := ids[item.ID]; ok {
			return newModelError(ErrDuplicateManagerID, fmt.Errorf("duplicate manager ID:%s", item.ID))
		}

		ids[item.ID] = true
	}

	return nil
}

func (this *EmployeeManagerCreateOption) Create(index SigningIndex) (
	[]dbmodels.CorporationManagerCreateOption, IModelError,
) {
	opt := make([]dbmodels.CorporationManagerCreateOption, 0, len(this.Managers))

	for i := range this.Managers {
		item := &this.Managers[i]

		v := dbmodels.CorporationManagerCreateOption{
			ID:       managerAccount(item.ID, item.Email),
			Name:     item.Name,
			Email:    item.Email,
			Role:     dbmodels.RoleManager,
			Password: util.RandStr(8, "alphanum"),
		}

		if err := dbmodels.GetDB().AddEmployeeManager(&index, &v); err != nil {
			if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
				return nil, newModelError(ErrNoLink, err)
			}
			return nil, parseDBError(err)
		}

		opt = append(opt, v)
	}

	return opt, nil
}

func (this *EmployeeManagerCreateOption) ValidateWhenDeleting(detail *dbmodels.CorporationDetail) IModelError {
	if len(this.Managers) == 0 {
		return newModelError(ErrEmptyPayload, fmt.Errorf("no employee mangers"))
	}

	for i := range this.Managers {
		item := &this.Managers[i]

		if err := checkEmailFormat(item.Email); err != nil {
			return err
		}

		if !detail.HasDomain(util.EmailSuffix(item.Email)) {
			return newModelError(ErrNotSameCorp, fmt.Errorf("not same email domain"))
		}

		if item.Email == detail.AdminEmail() {
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
