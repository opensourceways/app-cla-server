package pdf

import (
	"os"

	"github.com/opensourceways/gofpdf"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type IPDFGenerator interface {
	GenPDFForCorporationSigning(linkID, claFile string, signing *models.CorporationSigning, claFields []models.CLAField) (string, error)
}

var generator *pdfGenerator

func InitPDFGenerator(cfg *Config) error {
	path := util.GenFilePath(cfg.PDFOutDir, "tmp")
	if util.IsNotDir(path) {
		if err := os.Mkdir(path, 0644); err != nil {
			return err
		}
	}

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

		contactFont: fontInfo{font: "NotoSansSC-Regular", size: 10},

		seal:          "Seal",
		signature:     "Signature of Legal/Authorized Representative",
		signatureDate: "Date",

		newPDF: func() *gofpdf.Fpdf {
			pdf := gofpdf.New("P", "mm", "A4", "./conf/pdf-font") // 210mm x 297mm
			pdf.AddUTF8Font("NotoSansSC-Regular", "", "NotoSansSC-Regular.ttf")
			return pdf
		},
	}, nil
}

func newGeneratorForChinese() (*corpSigningPDF, error) {
	lang := "chinese"

	return &corpSigningPDF{
		language: lang,
		gh:       5.0,

		contactFont: fontInfo{font: "NotoSansSC-Regular", size: 10},

		seal:          "盖章",
		signature:     "法定/授权代表签字",
		signatureDate: "日期",

		newPDF: func() *gofpdf.Fpdf {
			pdf := gofpdf.New("P", "mm", "A4", "./conf/pdf-font") // 210mm x 297mm
			pdf.AddUTF8Font("NotoSansSC-Regular", "", "NotoSansSC-Regular.ttf")
			pdf.AddUTF8Font("NotoSansSC-Regular", "I", "NotoSansSC-Regular.ttf")
			return pdf
		},
	}, nil
}
