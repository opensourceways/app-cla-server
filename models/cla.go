package models

import "github.com/opensourceways/app-cla-server/dbmodels"

type CLA dbmodels.CLA

func (this *CLA) get(onlyFields bool) error {
	v, err := dbmodels.GetDB().GetCLA(this.ID, onlyFields)
	if err != nil {
		return err
	}
	*((*dbmodels.CLA)(this)) = v
	return err
}

func (this *CLA) Get() error {
	return this.get(false)
}

func (this *CLA) GetFields() error {
	return this.get(true)
}

func (this *CLA) Delete() error {
	return dbmodels.GetDB().DeleteCLA(this.ID)
}

type CLAListOptions dbmodels.CLAListOptions

func (this CLAListOptions) Get() ([]dbmodels.CLA, error) {
	return dbmodels.GetDB().ListCLA(dbmodels.CLAListOptions(this))
}

func ListCLAByIDs(ids []string) ([]dbmodels.CLA, error) {
	return dbmodels.GetDB().ListCLAByIDs(ids)
}
