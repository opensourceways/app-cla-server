package pdf

import (
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

func (this *corpSigningPDF) end(pdf *gofpdf.Fpdf, path string) error {
	if pdf.Err() {
		return fmt.Errorf("Failed to geneate pdf: %s", pdf.Error().Error())
	}

	return pdf.OutputFileAndClose(path)
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

func (c *corpSigningPDF) addSignature(pdf *gofpdf.Fpdf, items map[string]string, orders []string, titles map[string]string) {
	f := func(title, value string, border bool) {
		addItem(pdf, c.gh, title, value, border)
	}

	setFont(pdf, c.contactFont)
	multlines(pdf, c.gh, "")
	blankLine(pdf, 1)

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
