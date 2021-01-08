package pdf

import (
	"fmt"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type IPDFGenerator interface {
	LangSupported() map[string]bool
	GetBlankSignaturePath(string) string

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

	blankPDF := generator.GetBlankSignaturePath(c.language)
	if err = c.genBlankSignaturePage(blankPDF); err != nil {
		return err
	}
	return nil
}

func GetPDFGenerator() IPDFGenerator {
	return generator
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
