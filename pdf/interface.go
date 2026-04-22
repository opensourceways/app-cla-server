package pdf

import (
	"github.com/opensourceways/gofpdf"

	"github.com/opensourceways/app-cla-server/models"
)

const (
	fontNameNotoSansSC = "NotoSansSC-Regular"
	fontFileNotoSansSC = "NotoSansSC-Regular.ttf"
)

type IPDFGenerator interface {
	GenPDFForCorporationSigning(linkID, claFile string, signing *models.CorporationSigning, claFields []models.CLAField) (string, error)
}

var generator *pdfGenerator

func InitPDFGenerator(cfg *Config) error {
	generator = &pdfGenerator{
		pythonBin: cfg.PythonBin,
		pdfOutDir: cfg.PDFOutDir,
	}

	corp := []*corpSigningPDF{}
	m := []func() (*corpSigningPDF, error){
		newGeneratorForEnglish,
		newGeneratorForChinese,
	}
	for _, f := range m {
		c, err := f()
		if err != nil {
			return err
		}

		corp = append(corp, c)
	}

	generator.corp = corp
	return nil
}

func GetPDFGenerator() IPDFGenerator {
	return generator
}

func newGeneratorForEnglish() (*corpSigningPDF, error) {
	lang := "english"

	return &corpSigningPDF{
		language: lang,
		gh:       5.0,

		contactFont: fontInfo{font: fontNameNotoSansSC, size: 10},

		seal:          "Seal",
		signature:     "Signature of Legal/Authorized Representative",
		signatureDate: "Date",

		newPDF: func() *gofpdf.Fpdf {
			pdf := gofpdf.New("P", "mm", "A4", "./conf/pdf-font")
			pdf.AddUTF8Font(fontNameNotoSansSC, "", fontFileNotoSansSC)
			return pdf
		},
	}, nil
}

func newGeneratorForChinese() (*corpSigningPDF, error) {
	lang := "chinese"

	return &corpSigningPDF{
		language: lang,
		gh:       5.0,

		contactFont: fontInfo{font: fontNameNotoSansSC, size: 10},

		seal:          "盖章",
		signature:     "法定/授权代表签字",
		signatureDate: "日期",

		newPDF: func() *gofpdf.Fpdf {
			pdf := gofpdf.New("P", "mm", "A4", "./conf/pdf-font")
			pdf.AddUTF8Font(fontNameNotoSansSC, "", fontFileNotoSansSC)
			pdf.AddUTF8Font(fontNameNotoSansSC, "I", fontFileNotoSansSC)
			return pdf
		},
	}, nil
}