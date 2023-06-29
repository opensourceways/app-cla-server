package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
)

func NewCorpPDFAdapter(s app.CorpPDFService) *corpPDFAdatper {
	return &corpPDFAdatper{s}
}

type corpPDFAdatper struct {
	s app.CorpPDFService
}

func (adapter *corpPDFAdatper) Upload(csId string, pdf []byte) models.IModelError {
	if err := adapter.s.Upload(csId, pdf); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *corpPDFAdatper) Download(csId string) ([]byte, models.IModelError) {
	v, err := adapter.s.Download(csId)
	if err != nil {
		return nil, toModelError(err)
	}

	return v, nil
}
