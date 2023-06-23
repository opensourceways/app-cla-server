package domain

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

type AllSingingInfo = map[string]string

type Representative struct {
	Name      dp.Name
	EmailAddr dp.EmailAddr
}

type Link struct {
	Id       string
	CLAId    string
	Language dp.Language
}

type CorpSigning struct {
	Id      string
	PDF     string
	Date    string
	Link    Link
	Rep     Representative
	Corp    Corporation
	AllInfo AllSingingInfo
	Version int
}

func (cs *CorpSigning) AllEmailDomains() []string {
	return cs.Corp.AllEmailDomains
}
