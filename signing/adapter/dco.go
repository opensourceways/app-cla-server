package adapter

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/util"
)

func NewDCOAdapter(
	s app.DCOService,
	maxSizeOfDCOContent int,
	fileTypeOfDCOContent string,
) *dcoAdapter {
	return &dcoAdapter{
		s:                    s,
		maxSizeOfDCOContent:  maxSizeOfDCOContent,
		fileTypeOfDCOContent: fileTypeOfDCOContent,
	}
}

type dcoAdapter struct {
	s                    app.DCOService
	maxSizeOfDCOContent  int
	fileTypeOfDCOContent string
}

// Remove
func (adapter *dcoAdapter) Remove(linkId, claId string) models.IModelError {
	err := adapter.s.Remove(domain.CLAIndex{
		LinkId: linkId,
		CLAId:  claId,
	})

	if err != nil {
		return toModelError(err)
	}

	return nil
}

// List
func (adapter *dcoAdapter) List(linkId string) ([]models.CLADetail, models.IModelError) {
	v, err := adapter.s.List(linkId)
	if err != nil {
		return nil, toModelError(err)
	}

	return toCLADetail(v), nil
}

// DCOLocalFilePath
func (adapter *dcoAdapter) DCOLocalFilePath(linkId, claId string) string {
	return adapter.s.DCOLocalFilePath(domain.CLAIndex{
		LinkId: linkId,
		CLAId:  claId,
	})
}

// Add
func (adapter *dcoAdapter) Add(linkId string, opt *models.DCOCreateOpt) models.IModelError {
	cmd, err := adapter.cmdToAddDCO(opt)
	if err != nil {
		return errBadRequestParameter(err)
	}

	if err := adapter.s.Add(linkId, &cmd); err != nil {
		return toModelError(err)
	}

	return nil
}

func (adapter *dcoAdapter) cmdToAddDCO(opt *models.DCOCreateOpt) (
	cmd app.CmdToAddDCO, err error,
) {
	cmd.Text, err = util.DownloadFile(
		opt.URL, adapter.fileTypeOfDCOContent, adapter.maxSizeOfDCOContent,
	)
	if err != nil {
		return
	}

	if cmd.URL, err = dp.NewURL(opt.URL); err != nil {
		return
	}

	if cmd.Language, err = dp.NewLanguage(opt.Language); err != nil {
		return
	}

	cmd.Fields, err = toCLAFields(dp.CLATypeIndividual, opt.Fields)

	return
}
