package adapter

import (
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
)

func NewCorpAdminAdapter(s app.CorpAdminService) *corpAdminAdatper {
	return &corpAdminAdatper{s}
}

type corpAdminAdatper struct {
	s app.CorpAdminService
}

func (adapter *corpAdminAdatper) Add(csId string) (
	dbmodels.CorporationManagerCreateOption, models.IModelError,
) {
	dto, err := adapter.s.Add(csId)
	if err != nil {
		return dbmodels.CorporationManagerCreateOption{}, toModelError(err)
	}

	return dbmodels.CorporationManagerCreateOption{
		ID:       dto.Id,
		Role:     dbmodels.RoleAdmin,
		Name:     dto.Name,
		Email:    dto.EmailAddr,
		Password: dto.Password,
	}, nil
}
