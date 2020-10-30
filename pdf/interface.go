package pdf

import (
	"fmt"
	"io/ioutil"

	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
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
	path := "./conf/blank_signature/english_blank_signature.pdf"
	err := generator.corp.genBlankSignaturePage(path)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Failed to update blank signature: %s", err.Error())
	}

	err = dbmodels.GetDB().UploadBlankSignature(generator.corp.language, data)
	if err != nil {
		return fmt.Errorf("Failed to update blank signature: %s", err.Error())
	}
	return nil
}

func newCorpSigningPDF() (*corpSigningPDF, error) {
	path := "./conf/pdf_template_corporation/welcome.tmpl"
	welTemp, err := util.NewTemplate("wel", path)
	if err != nil {
		return nil, err
	}

	path = "./conf/pdf_template_corporation/declaration.tmpl"
	declTemp, err := util.NewTemplate("decl", path)
	if err != nil {
		return nil, err
	}

	return &corpSigningPDF{
		language:    "english",
		welcomeTemp: welTemp,
		declaration: declTemp,
		gh:          5.0,

		footerFont:    fontInfo{font: "Arial", size: 8},
		titleFont:     fontInfo{font: "Arial", size: 12},
		welcomeFont:   fontInfo{font: "Times", size: 12},
		contactFont:   fontInfo{font: "NotoSansSC-Regular", size: 12},
		declareFont:   fontInfo{font: "Times", size: 12},
		claFont:       fontInfo{font: "Times", size: 12},
		urlFont:       fontInfo{font: "Times", size: 12},
		signatureFont: fontInfo{font: "Arial", size: 12},

		subtitle: "Software Grant and Corporate Contributor License Agreement (\"Agreement\")",

		footerNumber: func(num int) string { return fmt.Sprintf("Page %d", num) },

		signatureItems: [][]string{
			{"Community Sign", "Corporation Sign"},
			{"Signature", "Signature"},
			{"Title", "Title"},
			{"Community", "Corporation"},
		},
		signatureDate: "Date",
	}, nil
}
