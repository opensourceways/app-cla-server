package adapter

import (
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
	models.CorporationManagerCreateOption, models.IModelError,
) {
	dto, err := adapter.s.Add(csId)
	if err != nil {
		return models.CorporationManagerCreateOption{}, toModelError(err)
	}

	return toCorporationManagerCreateOption(&dto), nil
}
