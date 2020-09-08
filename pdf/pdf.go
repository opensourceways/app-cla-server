package pdf

import (
	"fmt"
	"io/ioutil"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
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
	welTemp, err := newTemplate("wel", welcome)
	if err != nil {
		return err
	}

	declTemp, err := newTemplate("wel", declPath)
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

func UploadBlankSignature(language, path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Failed to update blank siganture: %s", err.Error())
	}

	err = dbmodels.GetDB().UploadBlankSignature(language, data)
	if err != nil {
		return fmt.Errorf("Failed to update blank siganture: %s", err.Error())
	}
	return nil
}
