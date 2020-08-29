package models

import "github.com/zengchen1024/cla-server/dbmodels"

type IndividualSigning struct {
	CLAOrgID string                 `json:"cla_org_id"`
	Email    string                 `json:"email"`
	Info     map[string]interface{} `json:"info"`
}

func (this *IndividualSigning) Create() error {
	p := dbmodels.IndividualSigningInfo{}
	if err := copyBetweenStructs(this, &p); err != nil {
		return err
	}

	return dbmodels.GetDB().SignAsIndividual(this.CLAOrgID, p)
}
