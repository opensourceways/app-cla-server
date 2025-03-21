package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type CLAInfo struct {
	URL      dp.URL
	Text     []byte
	Type     dp.CLAType
	Fields   []domain.Field
	Language dp.Language
}

func (cmd *CLAInfo) toCLA() domain.CLA {
	return domain.CLA{
		URL:      cmd.URL,
		Text:     cmd.Text,
		Type:     cmd.Type,
		Fields:   cmd.Fields,
		Language: cmd.Language,
	}
}

type CmdToAddCLA struct {
	UserId string
	LinkId string

	CLAInfo
}

type CmdToRemoveCLA struct {
	UserId string

	domain.CLAIndex
}

type CLADTO struct {
	Id       string
	Type     string
	URL      string
	Language string
}

type CLADetailDTO struct {
	Id        string
	Fileds    []domain.Field
	Language  string
	LocalFile string
}
