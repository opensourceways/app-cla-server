package pdf

import (
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

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

func genSignatureItems(pdf *gofpdf.Fpdf, gh float64, ltips, rtips string, items [][]string) {
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	w := 92.5
	pdf.CellFormat(w, gh, ltips, "", 0, "C", false, 0, "")
	pdf.Cell(5, gh, "")
	pdf.CellFormat(w, gh, rtips, "", 1, "C", false, 0, "")
	pdf.Ln(10)

	for _, item := range items {
		addSignatureItem(pdf, gh, item[0], item[1], "", "")
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
