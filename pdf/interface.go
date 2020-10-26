package pdf

import (
	"fmt"
	"io/ioutil"

	"github.com/jung-kurt/gofpdf"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
)

type IPDFGenerator interface {
	GenPDFForCorporationSigning(orgCLA *models.OrgCLA, signing *models.CorporationSigning, cla *models.CLA) (string, error)
}

var generator *pdfGenerator

func InitPDFGenerator(pythonBin, pdfOutDir, pdfOrgSigDir string) error {
	c, err := newCorpSigningPDF()
	if err != nil {
		return err
	}
	generator = &pdfGenerator{
		pythonBin:    pythonBin,
		pdfOutDir:    pdfOutDir,
		pdfOrgSigDir: pdfOrgSigDir,
		corp:         c,
	}
	return nil
}

func GetPDFGenerator() IPDFGenerator {
	return generator
}

func GenBlankSignaturePage() error {
	corporation := &corpSigningPDF{gh: 5.0}

	pdf := gofpdf.New("P", "mm", "A4", "") // 210mm x 297mm

	items := [][]string{
		{"Signature", "Signature"},
		{"Title", "Title"},
		{"Community", "Corporation"},
	}

	genSignatureItems(pdf, corporation.gh, "Community Sign", "Corporation Sign", items)

	path := "./conf/blank_signature/english_blank_signature.pdf"
	if err := corporation.end(pdf, path); err != nil {
		return err
	}

	return uploadBlankSignature("english", path)
}

func uploadBlankSignature(language, path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Failed to update blank signature: %s", err.Error())
	}

	err = dbmodels.GetDB().UploadBlankSignature(language, data)
	if err != nil {
		return fmt.Errorf("Failed to update blank signature: %s", err.Error())
	}
	return nil
}
