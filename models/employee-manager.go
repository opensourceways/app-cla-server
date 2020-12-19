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

func (this *EmployeeManagerCreateOption) Validate(adminEmail string) *ModelError {
	if len(this.Managers) == 0 {
		return newModelError(ErrMissingParameter, fmt.Errorf("no employee mangers to add/delete"))
	}

	ids := map[string]bool{}
	em := map[string]bool{}
	suffix := util.EmailSuffix(adminEmail)

	for _, item := range this.Managers {
		if item.Email == "" {
			return newModelError(ErrMissingParameter, fmt.Errorf("missing email"))
		}

		if item.Email == adminEmail {
			return newModelError(ErrAddAdminAsManager, fmt.Errorf("can't add/delete administrator himself/herself"))
		}

		if merr := checkEmailFormat(item.Email); merr != nil {
			return merr
		}

		es := util.EmailSuffix(item.Email)
		if es != suffix {
			return newModelError(ErrNotSameCorp, fmt.Errorf("not same email suffix"))
		}

		if _, ok := em[item.Email]; ok {
			return newModelError(ErrDuplicateManagerEmail, fmt.Errorf("duplicate email:%s", item.Email))
		}
		em[item.Email] = true

		if item.ID != "" {
			if err := checkManagerID(fmt.Sprintf("%s_%s", item.ID, es)); err != nil {
				return err
			}

			if _, ok := ids[item.ID]; ok {
				return newModelError(ErrDuplicateManagerID, fmt.Errorf("duplicate manager ID:%s", item.ID))
			}
			ids[item.ID] = true
		}
	}

	return nil
}

func (this *EmployeeManagerCreateOption) Create(linkID string) ([]dbmodels.CorporationManagerCreateOption, *ModelError) {
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

	err := dbmodels.GetDB().AddCorporationManager(linkID, opt, conf.AppConfig.EmployeeManagersNumber)
	if err == nil {
		es := util.EmailSuffix(opt[0].Email)
		for i := range opt {
			if opt[i].ID != "" {
				opt[i].ID = fmt.Sprintf("%s_%s", opt[i].ID, es)
			}
		}
		return opt, nil
	}

	if err.IsErrorOf(dbmodels.ErrMarshalDataFaield) {
		return nil, newModelError(ErrSystemError, err)
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return nil, newModelError(ErrNoLinkOrDuplicateManager, err)
	}

	return nil, parseDBError(err)

}

func (this *EmployeeManagerCreateOption) Delete(linkID string) ([]dbmodels.CorporationManagerCreateOption, *ModelError) {
	emails := make([]string, 0, len(this.Managers))

	for _, item := range this.Managers {
		emails = append(emails, item.Email)
	}

	v, err := dbmodels.GetDB().DeleteCorporationManager(linkID, emails)
	if err == nil {
		return v, nil
	}

	if err.IsErrorOf(dbmodels.ErrNoDBRecord) {
		return nil, newModelError(ErrNoLink, err)
	}

	return nil, parseDBError(err)
}
