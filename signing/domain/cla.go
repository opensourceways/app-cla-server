package domain

import "github.com/opensourceways/app-cla-server/signing/domain/dp"

type CLAIndex struct {
	LinkId string
	CLAId  string
}

type CLA struct {
	Id       string
	URL      dp.URL
	Text     []byte
	Type     dp.CLAType
	Fields   []Field
	Language dp.Language
}

func (cla *CLA) isMe(cla1 *CLA) bool {
	b := cla.Type.CLAType() == cla1.Type.CLAType()
	b1 := cla.Language.Language() == cla1.Language.Language()

	return b && b1
}

type Field struct {
	Id       string
	Required bool

	dp.CLAField
}
