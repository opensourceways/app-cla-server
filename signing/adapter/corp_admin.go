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

func (adapter *corpAdminAdatper) Add(userId, csId string) (
	string, models.CorporationManagerCreateOption, models.IModelError,
) {
	linkId, dto, err := adapter.s.Add(userId, csId)
	if err != nil {
		return "", models.CorporationManagerCreateOption{}, toModelError(err)
	}

	return linkId, toCorporationManagerCreateOption(&dto), nil
}
