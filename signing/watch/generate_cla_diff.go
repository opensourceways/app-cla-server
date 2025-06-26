package watch

import (
	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
)

func Instance() *watchingImpl {
	return impl
}

func (impl *watchingImpl) GenCLADiff(linkId, oldCLAId, newCLAId string) {
	oldPDFPath := models.CLAFile(linkId, oldCLAId)
	newPDFPath := models.CLAFile(linkId, newCLAId)
	diffFile := models.DiffCLAFile(linkId, oldCLAId, newCLAId)

	err := pdf.GetPDFGenerator().GenPDFDiff(oldPDFPath, newPDFPath, diffFile)
	if err != nil {
		logs.Error("generate diff pdf failed:", diffFile, err)
	}
}
