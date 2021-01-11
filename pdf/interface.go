package pdf

import (
	"fmt"
	"text/template"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type IPDFGenerator interface {
	LangSupported() map[string]bool
	GetBlankSignaturePath(string) string

	GenPDFForCorporationSigning(linkID, orgSignatureFile, claFile string, orgInfo *models.OrgInfo, signing *models.CorporationSigning, claFields []models.CLAField) (string, error)
}

var generator *pdfGenerator

func InitPDFGenerator(pythonBin, pdfOutDir, pdfOrgSigDir string) error {
	generator = &pdfGenerator{
		pythonBin:    pythonBin,
		pdfOutDir:    pdfOutDir,
		pdfOrgSigDir: pdfOrgSigDir,
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

		blankPDF := generator.GetBlankSignaturePath(c.language)
		if err = c.genBlankSignaturePage(blankPDF); err != nil {
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

	welTemp, err := newWelcomeTmpl(lang)
	if err != nil {
		return nil, err
	}

	declTemp, err := newDeclTmpl(lang)
	if err != nil {
		return nil, err
	}

	return &corpSigningPDF{
		language:    lang,
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

func newGeneratorForChinese() (*corpSigningPDF, error) {
	lang := "chinese"

	welTemp, err := newWelcomeTmpl(lang)
	if err != nil {
		return nil, err
	}

	declTemp, err := newDeclTmpl(lang)
	if err != nil {
		return nil, err
	}

	return &corpSigningPDF{
		language:    lang,
		welcomeTemp: welTemp,
		declaration: declTemp,
		gh:          5.0,

		footerFont:    fontInfo{font: "NotoSansSC-Regular", size: 8},
		titleFont:     fontInfo{font: "NotoSansSC-Regular", size: 12},
		welcomeFont:   fontInfo{font: "NotoSansSC-Regular", size: 12},
		contactFont:   fontInfo{font: "NotoSansSC-Regular", size: 12},
		declareFont:   fontInfo{font: "NotoSansSC-Regular", size: 12},
		claFont:       fontInfo{font: "NotoSansSC-Regular", size: 12},
		urlFont:       fontInfo{font: "Times", size: 12},
		signatureFont: fontInfo{font: "NotoSansSC-Regular", size: 12},

		subtitle: "软件授权和企业贡献者许可协议 (\"协议\")",

		footerNumber: func(num int) string { return fmt.Sprintf("%d 页", num) },

		signatureItems: [][]string{
			{"社区签署", "企业签署"},
			{"签名", "签名"},
			{"职位", "职位"},
			{"社区名称", "企业名称"},
		},
		signatureDate: "日期",
	}, nil
}

func newWelcomeTmpl(lang string) (*template.Template, error) {
	return util.NewTemplate("wel", pathOfTemplate("welcome", lang))
}

func newDeclTmpl(lang string) (*template.Template, error) {
	return util.NewTemplate("decl", pathOfTemplate("declaration", lang))
}

func pathOfTemplate(part, lang string) string {
	return fmt.Sprintf("./conf/pdf_template_corporation/%s_%s.tmpl", part, lang)
}
