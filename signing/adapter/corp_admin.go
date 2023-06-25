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

func (adapter *corpAdminAdatper) Add(csId string) models.IModelError {
	if err := adapter.s.Add(csId); err != nil {
		return toModelError(err)
	}

	return nil
}
