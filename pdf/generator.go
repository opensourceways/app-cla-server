package pdf

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/jung-kurt/gofpdf"

	"github.com/opensourceways/app-cla-server/util"
)

type corporationCLAPDF struct {
	welcomeTemp *template.Template
	declaration *template.Template
	gh          float64
}

func newCorporationPDF() (*corporationCLAPDF, error) {
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

	return &corporationCLAPDF{
		welcomeTemp: welTemp,
		declaration: declTemp,
		gh:          5.0,
	}, nil
}

func (this *corporationCLAPDF) begin() *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "./conf/pdf-font") // 210mm x 297mm
	pdf.AddUTF8Font("NotoSansSC-Regular", "", "NotoSansSC-Regular.ttf")
	initializePdf(pdf)
	return pdf
}

func (this *corporationCLAPDF) end(pdf *gofpdf.Fpdf, path string) error {
	if pdf.Err() {
		return fmt.Errorf("Failed to geneate pdf: %s", pdf.Error().Error())
	}

	return pdf.OutputFileAndClose(path)
}

func (this *corporationCLAPDF) firstPage(pdf *gofpdf.Fpdf, title string) {
	pdf.AddPage()

	pdf.SetFont("Arial", "", 12)

	pdf.CellFormat(0, 10, title, "", 1, "C", false, 0, "")

	desc := "Software Grant and Corporate Contributor License Agreement (\"Agreement\")"
	pdf.CellFormat(0, 5, desc, "", 1, "C", false, 0, "")

	pdf.Ln(-1)
}

func (this *corporationCLAPDF) welcome(pdf *gofpdf.Fpdf, project, email string) {
	data := struct {
		Project string
		Email   string
	}{
		Project: project,
		Email:   email,
	}

	tmpl := this.welcomeTemp

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		pdf.SetErrorf("Failed to add welcome part: execute template failed: %s", err.Error())
		return
	}

	multlines(pdf, this.gh, buf.String())
}

func (this *corporationCLAPDF) contact(pdf *gofpdf.Fpdf, items map[string]string, orders []string, keys map[string]string) {
	gh := this.gh

	f := func(title, value string) {
		pdf.CellFormat(50, gh, fmt.Sprintf("%s:", title), "", 0, "R", false, 0, "")

		pdf.Cell(2, gh, " ")

		pdf.MultiCell(130, gh, value, "B", "L", false)
	}

	defer func() {
		pdf.SetFont("Arial", "", 12)
	}()

	pdf.SetFont("NotoSansSC-Regular", "", 12)

	for _, i := range orders {
		f(keys[i], items[i])
		pdf.Ln(-1)
	}
}

func (this *corporationCLAPDF) declare(pdf *gofpdf.Fpdf, project string) {
	data := struct {
		Project string
	}{
		Project: project,
	}

	tmpl := this.declaration

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		pdf.SetErrorf("Failed to add declaration part: execute template failed: %s", err.Error())
		return
	}

	multlines(pdf, this.gh, buf.String())
}

func (this *corporationCLAPDF) cla(pdf *gofpdf.Fpdf, content string) {
	multlines(pdf, this.gh, content)
}

func (this *corporationCLAPDF) secondPage(pdf *gofpdf.Fpdf, date string) {
	item := []string{"", ""}
	items := [][]string{item, item, item}
	genSignatureItems(pdf, this.gh, "", "", items)

	addSignatureItem(pdf, this.gh, "Date", "Date", date, "")
}
