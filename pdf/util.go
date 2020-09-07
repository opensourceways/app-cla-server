package pdf

import (
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/jung-kurt/gofpdf"
)

func addSignatureItem(pdf *gofpdf.Fpdf, gh float64, title, value string) {
	b := ""
	if title != "" {
		b = "B"
	}

	w := 92.5
	pdf.Cell(w, gh, title)
	pdf.Cell(5, gh, "")
	pdf.CellFormat(w, gh, title, "", 1, "L", false, 0, "")

	pdf.Ln(-1)

	pdf.CellFormat(w, gh, value, b, 0, "L", false, 0, "")
	pdf.Cell(5, gh, "")
	pdf.CellFormat(w, gh, value, b, 1, "L", false, 0, "")

	pdf.Ln(-1)
}

func signature(pdf *gofpdf.Fpdf, gh float64, guidances string, items []string) {
	pdf.SetFont("Arial", "", 12)

	b := ""
	if guidances != "" {
		b = "B"
	}

	pdf.CellFormat(0, 10, guidances, b, 1, "", false, 0, "")

	pdf.Ln(5)

	for _, item := range items {
		addSignatureItem(pdf, gh, item, "")
	}
}

func multlines(pdf *gofpdf.Fpdf, gh float64, content string) {
	// Times 12
	pdf.SetFont("Times", "", 12)
	// Output justified text
	pdf.MultiCell(0, gh, content, "", "", false)
	// Line break
	pdf.Ln(-1)
}

func initializePdf(pdf *gofpdf.Fpdf) {
	pdf.SetFooterFunc(func() {
		// Position at 1.5 cm from bottom
		pdf.SetY(-15)
		// Arial italic 8
		pdf.SetFont("Arial", "I", 8)
		// Text color in gray
		pdf.SetTextColor(128, 128, 128)
		// Page number
		pdf.CellFormat(
			0, 10, fmt.Sprintf("Page %d", pdf.PageNo()),
			"", 0, "C", false, 0, "",
		)
	})
}

func newTemplate(name, path string) (*template.Template, error) {
	txtStr, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to new template: read template file failed: %s", err.Error())
	}

	tmpl, err := template.New(name).Parse(string(txtStr))
	if err != nil {
		return nil, fmt.Errorf("Failed to new template: build template failed: %s", err.Error())
	}

	return tmpl, nil
}
