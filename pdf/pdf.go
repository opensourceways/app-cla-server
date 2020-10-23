package pdf

import (
	"github.com/opensourceways/app-cla-server/models"
)

type IPDFGenerator interface {
	GenCLAPDFForCorporation(orgCLA *models.OrgCLA, signing *models.CorporationSigning, cla *models.CLA) (string, error)
}

var generator *pdfGenerator

type pdfGenerator struct {
	pdfOutDir    string
	pdfOrgSigDir string
	pythonBin    string
	corporation  *corporationCLAPDF
}

func InitPDFGenerator(pythonBin, pdfOutDir, pdfOrgSigDir string) error {
	c, err := newCorporationPDF()
	if err != nil {
		return err
	}
	generator = &pdfGenerator{
		pythonBin:    pythonBin,
		pdfOutDir:    pdfOutDir,
		pdfOrgSigDir: pdfOrgSigDir,
		corporation:  c,
	}
	return nil
}

func GetPDFGenerator() IPDFGenerator {
	return generator
}
