package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewCorpPDFAdapter(s app.CorpPDFService) *corpPDFAdatper {
	return &corpPDFAdatper{s}
}

type corpPDFAdatper struct {
	s app.CorpPDFService
}

func (adapter *corpPDFAdatper) Upload(userId, csId string, pdf []byte) models.IModelError {
	cmd := app.CmdToUploadCorpSigningPDF{
		UserId: userId,
		CSId:   csId,
		PDF:    pdf,
	}

	if err := adapter.s.Upload(&cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *corpPDFAdatper) Download(userId, csId string, email dp.EmailAddr) ([]byte, models.IModelError) {
	v, err := adapter.s.Download(userId, csId, email)
	if err != nil {
		return nil, toModelError(err)
	}

	return v, nil
}
