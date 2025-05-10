/*
 * Copyright (C) 2021. Huawei Technologies Co., Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pdf

import (
	"fmt"
	"text/template"

	"github.com/opensourceways/gofpdf"

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
			{"Signature", "Signature and Seal"},
			{"Title", "Title"},
			{"Community", "Corporation"},
		},
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
			{"签名", "签名(加盖公章)"},
			{"职位", "职位"},
			{"社区名称", "企业名称"},
		},
		signatureDate: "日期",

		newPDF: func() *gofpdf.Fpdf {
			pdf := gofpdf.New("P", "mm", "A4", "./conf/pdf-font") // 210mm x 297mm
			pdf.AddUTF8Font("NotoSansSC-Regular", "", "NotoSansSC-Regular.ttf")
			pdf.AddUTF8Font("NotoSansSC-Regular", "I", "NotoSansSC-Regular.ttf")
			return pdf
		},
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
