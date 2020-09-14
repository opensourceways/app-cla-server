package pdf

import (
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type IPDFGenerator interface {
	GenCLAPDFForCorporation(claOrg *models.CLAOrg, signing *models.CorporationSigning, cla *models.CLA) (string, error)
}

var generator *pdfGenerator

type pdfGenerator struct {
	pdfOutDir    string
	pdfOrgSigDir string
	pythonBin    string
	corporation  *corporationCLAPDF
}

func InitPDFGenerator(pythonBin, pdfOutDir, pdfOrgSigDir, welcome, declPath string) error {
	welTemp, err := util.NewTemplate("wel", welcome)
	if err != nil {
		return err
	}

	declTemp, err := util.NewTemplate("wel", declPath)
	if err != nil {
		return err
	}

	generator = &pdfGenerator{
		pythonBin:    pythonBin,
		pdfOutDir:    pdfOutDir,
		pdfOrgSigDir: pdfOrgSigDir,
		corporation: &corporationCLAPDF{
			welcomeTemp: welTemp,
			declaration: declTemp,
			gh:          5.0,
		},
	}
	return nil
}

func GetPDFGenerator() IPDFGenerator {
	return generator
}
