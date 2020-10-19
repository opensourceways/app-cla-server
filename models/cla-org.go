package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type CLAOrg dbmodels.CLAOrg

func (this *CLAOrg) Create() error {
	this.Enabled = true

	v, err := dbmodels.GetDB().CreateBindingBetweenCLAAndOrg(*(*dbmodels.CLAOrg)(this))
	if err == nil {
		this.ID = v
	}

	return err
}

func (this CLAOrg) Delete() error {
	return dbmodels.GetDB().DeleteBindingBetweenCLAAndOrg(this.ID)
}

func (this *CLAOrg) Get() error {
	v, err := dbmodels.GetDB().GetBindingBetweenCLAAndOrg(this.ID)
	if err != nil {
		return err
	}
	*(*dbmodels.CLAOrg)(this) = v
	return nil
}

type CLAOrgListOption dbmodels.CLAOrgListOption

func (this CLAOrgListOption) List() ([]dbmodels.CLAOrg, error) {
	return dbmodels.GetDB().ListBindingBetweenCLAAndOrg(dbmodels.CLAOrgListOption(this))
}
