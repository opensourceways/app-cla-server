package pdf

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/opensourceways/gofpdf"
)

type fontInfo struct {
	font string
	size float64
}

type corpSigningPDF struct {
	language string

	welcomeTemp *template.Template
	declaration *template.Template
	gh          float64

	footerFont    fontInfo
	titleFont     fontInfo
	welcomeFont   fontInfo
	contactFont   fontInfo
	declareFont   fontInfo
	claFont       fontInfo
	urlFont       fontInfo
	signatureFont fontInfo

	subtitle     string
	footerNumber func(int) string

	signatureItems [][]string
	seal           string
	signature      string
	signatureDate  string
	newPDF         func() *gofpdf.Fpdf
}

func (this *corpSigningPDF) begin() *gofpdf.Fpdf {
	pdf := this.newPDF()

	pdf.SetFooterFunc(func() {
		// Position at 1.5 cm from bottom
		pdf.SetY(-15)
		// Arial italic 8
		pdf.SetFont(this.footerFont.font, "I", this.footerFont.size)
		// Text color in gray
		pdf.SetTextColor(128, 128, 128)
		// Page number
		pdf.CellFormat(
			0, 10, this.footerNumber(pdf.PageNo()),
			"", 0, "C", false, 0, "",
		)
	})

	return pdf
}

func (this *corpSigningPDF) end(pdf *gofpdf.Fpdf, path string) error {
	if pdf.Err() {
		return fmt.Errorf("Failed to geneate pdf: %s", pdf.Error().Error())
	}

	return pdf.OutputFileAndClose(path)
}

func (this *corpSigningPDF) firstPage(pdf *gofpdf.Fpdf, title string) {
	pdf.AddPage()

	setFont(pdf, this.titleFont)

	pdf.CellFormat(0, 10, title, "", 1, "C", false, 0, "")

	pdf.CellFormat(0, 5, this.subtitle, "", 1, "C", false, 0, "")

	pdf.Ln(-1)
}

func (this *corpSigningPDF) welcome(pdf *gofpdf.Fpdf, org, email string) {
	data := struct {
		Org   string
		Email string
	}{
		Org:   org,
		Email: email,
	}

	tmpl := this.welcomeTemp

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		pdf.SetErrorf("Failed to add welcome part: execute template failed: %s", err.Error())
		return
	}

	setFont(pdf, this.welcomeFont)
	multlines(pdf, this.gh, buf.String())
}

func addItem(pdf *gofpdf.Fpdf, gh float64, title, value string, needBorder bool) {
	w1 := 32.0
	w := 210 - 2*w1

	// the default blank space is 10mm
	w1 -= 10
	pdf.Cell(w1, gh, "")
	pdf.CellFormat(w, gh, fmt.Sprintf("%s:", title), "", 1, "L", false, 0, "")
	blankLine(pdf, 1)

	b := ""
	if needBorder {
		b = "B"
	}

	pdf.Cell(w1, gh, "")
	pdf.CellFormat(w, gh, "    "+value, b, 1, "L", false, 0, "")
	blankLine(pdf, 2)
}

func (this *corpSigningPDF) contact(pdf *gofpdf.Fpdf, items map[string]string, orders []string, titles map[string]string) {
	gh := this.gh

	f := func(title, value string) {
		pdf.CellFormat(50, gh, fmt.Sprintf("%s:", title), "", 0, "R", false, 0, "")

		pdf.Cell(2, gh, " ")

		pdf.MultiCell(130, gh, value, "B", "L", false)

		pdf.Ln(-1)
	}

	setFont(pdf, this.contactFont)

	for _, i := range orders {
		f(titles[i], items[i])
	}
}

func (this *corpSigningPDF) declare(pdf *gofpdf.Fpdf) {
	tmpl := this.declaration

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, nil); err != nil {
		pdf.SetErrorf("Failed to add declaration part: execute template failed: %s", err.Error())
		return
	}

	setFont(pdf, this.declareFont)
	multlines(pdf, this.gh, buf.String())
}

func (this *corpSigningPDF) cla(pdf *gofpdf.Fpdf, content string) {
	setFont(pdf, this.claFont)
	multlines(pdf, this.gh, content)
}

func (this *corpSigningPDF) projectURL(pdf *gofpdf.Fpdf, url string) {
	setFont(pdf, this.urlFont)
	multlines(pdf, this.gh, url)
}

func (this *corpSigningPDF) secondPage(pdf *gofpdf.Fpdf, date string) {
	items := make([][]string, len(this.signatureItems))
	for i := range items {
		items[i] = []string{"", ""}
	}

	this.genSignatureItems(pdf, items)

	addSignatureItem(pdf, this.gh, this.signatureDate, this.signatureDate, date, "")
}

func (this *corpSigningPDF) genBlankSignaturePage(path string) error {
	pdf := this.newPDF()

	this.genSignatureItems(pdf, this.signatureItems)

	return this.end(pdf, path)
}

func (this *corpSigningPDF) genSignatureItems(pdf *gofpdf.Fpdf, items [][]string) {
	pdf.AddPage()
	setFont(pdf, this.signatureFont)

	w := 92.5
	gh := this.gh

	pdf.CellFormat(w, gh, items[0][0], "", 0, "C", false, 0, "")
	pdf.Cell(5, gh, "")

	pdf.CellFormat(w, gh, items[0][1], "", 1, "C", false, 0, "")
	pdf.Ln(10)

	for i := 1; i < len(items); i++ {
		addSignatureItem(pdf, gh, items[i][0], items[i][1], "", "")
	}
}

func (c *corpSigningPDF) addSignature(pdf *gofpdf.Fpdf, items map[string]string, orders []string, titles map[string]string) {
	f := func(title, value string, border bool) {
		addItem(pdf, c.gh, title, value, border)
	}

	setFont(pdf, c.contactFont)

	for _, i := range orders {
		f(titles[i], items[i], true)
	}

	f(c.signature, "", true)

	w1 := 32.0
	w := 105 - w1
	// the default blank space is 10mm
	w1 -= 10
	gh := c.gh

	pdf.Cell(w1, gh, "")
	pdf.Cell(w, gh, c.seal+":")
	pdf.CellFormat(w, gh, c.signatureDate+":", "", 1, "L", false, 0, "")
}

func addSignatureItem(pdf *gofpdf.Fpdf, gh float64, ltitle, rtitle, lvalue, rvalue string) {
	w := 92.5

	b := ""
	if ltitle != "" {
		b = "B"
	}

	pdf.Cell(w, gh, ltitle)
	pdf.Cell(5, gh, "")
	pdf.CellFormat(w, gh, rtitle, "", 1, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(w, gh, lvalue, b, 0, "L", false, 0, "")
	pdf.Cell(5, gh, "")
	pdf.CellFormat(w, gh, rvalue, b, 1, "L", false, 0, "")
	pdf.Ln(-1)
}

func multlines(pdf *gofpdf.Fpdf, gh float64, content string) {
	// Output justified text
	pdf.MultiCell(0, gh, content, "", "", false)
	// Line break
	pdf.Ln(-1)
}

func setFont(pdf *gofpdf.Fpdf, font fontInfo) {
	pdf.SetFont(font.font, "", font.size)
}

func blankLine(pdf *gofpdf.Fpdf, n int) {
	for i := 0; i < n; i++ {
		pdf.Ln(-1)
	}
}
