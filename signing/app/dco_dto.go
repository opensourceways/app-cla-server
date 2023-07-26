package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type CmdToAddDCO struct {
	URL      dp.URL
	Text     []byte
	Fields   []domain.Field
	Language dp.Language
}

func (cmd *CmdToAddDCO) toDCO() domain.CLA {
	return domain.CLA{
		URL:      cmd.URL,
		Text:     cmd.Text,
		Type:     dp.CLATypeIndividual,
		Fields:   cmd.Fields,
		Language: cmd.Language,
	}
}
