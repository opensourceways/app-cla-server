package models

import (
	"fmt"

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

func (this *EmployeeManagerCreateOption) ValidateWhenDeleting(adminEmail string, emailDomains map[string]bool) IModelError {
	if len(this.Managers) == 0 {
		return newModelError(ErrEmptyPayload, fmt.Errorf("no employee mangers"))
	}

	for i := range this.Managers {
		item := &this.Managers[i]

		if err := checkEmailFormat(item.Email); err != nil {
			return err
		}

		if !emailDomains[util.EmailSuffix(item.Email)] {
			return newModelError(ErrNotSameCorp, fmt.Errorf("not same email domain"))
		}

		if item.Email == adminEmail {
			return newModelError(ErrAdminAsManager, fmt.Errorf("can't delete administrator"))
		}
	}

	return nil
}
