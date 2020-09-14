package models

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type IndividualSigning struct {
	Email string                   `json:"email"`
	Name  string                   `json:"name"`
	Info  dbmodels.TypeSigningInfo `json:"info"`
}

func (this *IndividualSigning) Create(claOrgID string) error {
	p := dbmodels.IndividualSigningInfo{}
	if err := util.CopyBetweenStructs(this, &p); err != nil {
		return err
	}
	p.Enabled = true

	return dbmodels.GetDB().SignAsIndividual(claOrgID, p)
}
