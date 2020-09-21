package models

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/util"
)

type IndividualSigning struct {
	Email   string                   `json:"email"`
	Name    string                   `json:"name"`
	Date    string                   `json:"date"`
	Enabled bool                     `json:"enabled"`
	Info    dbmodels.TypeSigningInfo `json:"info"`
}

func (this *IndividualSigning) Create(claOrgID string, enabled bool) error {
	this.Date = util.Date()
	this.Enabled = enabled
	p := dbmodels.IndividualSigningInfo{}
	if err := util.CopyBetweenStructs(this, &p); err != nil {
		return err
	}

	return dbmodels.GetDB().SignAsIndividual(claOrgID, p)
}

func IsIndividualSigned(platform, orgID, repoId, email string) (bool, error) {
	opt := dbmodels.IndividualSigningCheckInfo{
		Platform: platform,
		OrgID:    orgID,
		RepoID:   repoId,
		Email:    email,
	}

	return dbmodels.GetDB().IsIndividualSigned(opt)
}
