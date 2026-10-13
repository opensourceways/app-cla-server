package pdf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

func TestGenCLATemplateFileName(t *testing.T) {
	tests := []struct {
		linkID string
		other  string
		want   string
	}{
		{"link1", "_template", "link1_template.pdf"},
		{"link2", "_tmpl", "link2_tmpl.pdf"},
		{"", "_template", "_template.pdf"},
	}

	for _, tt := range tests {
		got := genCLATemplateFileName(tt.linkID, tt.other)
		if got != tt.want {
			t.Errorf("genCLATemplateFileName(%s, %s) = %s, want %s", tt.linkID, tt.other, got, tt.want)
		}
	}
}

func TestBuildCorpContactEmpty(t *testing.T) {
	orders, titles := BuildCorpContact([]models.CLAField{})

	if len(orders) != 0 {
		t.Errorf("expected 0 orders, got %d", len(orders))
	}
	if len(titles) != 0 {
		t.Errorf("expected 0 titles, got %d", len(titles))
	}
}

func TestBuildCorpContact(t *testing.T) {
	fields := []models.CLAField{
		{ID: "2", Title: "Name", Type: "name"},
		{ID: "1", Title: "Email", Type: "email"},
		{ID: "3", Title: "Phone", Type: "phone"},
	}

	orders, titles := BuildCorpContact(fields)

	if len(orders) != 3 {
		t.Fatalf("expected 3 orders, got %d", len(orders))
	}

	// orders should be sorted by numeric ID: 1, 2, 3
	if orders[0] != "1" || orders[1] != "2" || orders[2] != "3" {
		t.Errorf("orders = %v, want [1 2 3]", orders)
	}

	// titles map
	if titles["1"] != "Email" {
		t.Errorf("titles[1] = %s, want Email", titles["1"])
	}
	if titles["2"] != "Name" {
		t.Errorf("titles[2] = %s, want Name", titles["2"])
	}
	if titles["3"] != "Phone" {
		t.Errorf("titles[3] = %s, want Phone", titles["3"])
	}
}

func TestBuildCorpContactInvalidID(t *testing.T) {
	fields := []models.CLAField{
		{ID: "abc", Title: "Invalid", Type: "type"},
		{ID: "1", Title: "Valid", Type: "type"},
	}

	orders, _ := BuildCorpContact(fields)

	// Invalid ID "abc" should be skipped
	if len(orders) != 1 {
		t.Fatalf("expected 1 order (invalid skipped), got %d", len(orders))
	}
	if orders[0] != "1" {
		t.Errorf("orders[0] = %s, want 1", orders[0])
	}
}

func TestGenTemplateSignaturePDFEmpty(t *testing.T) {
	corp, err := newGeneratorForEnglish()
	if err != nil {
		t.Fatalf("newGeneratorForEnglish failed: %v", err)
	}

	outFile := filepath.Join(t.TempDir(), "test_sig_empty.pdf")
	err = genTemplateSignaturePDF(corp, []models.CLAField{}, outFile)
	if err != nil {
		t.Fatalf("genTemplateSignaturePDF with empty fields failed: %v", err)
	}
	if util.IsFileNotExist(outFile) {
		t.Error("expected output file to be created")
	}
}

func TestGenTemplateSignaturePDFWithFields(t *testing.T) {
	corp, err := newGeneratorForChinese()
	if err != nil {
		t.Fatalf("newGeneratorForChinese failed: %v", err)
	}

	fields := []models.CLAField{
		{ID: "1", Title: "Name", Type: "name"},
		{ID: "2", Title: "Email", Type: "email"},
	}

	outFile := filepath.Join(t.TempDir(), "test_sig_fields.pdf")
	err = genTemplateSignaturePDF(corp, fields, outFile)
	if err != nil {
		t.Fatalf("genTemplateSignaturePDF with fields failed: %v", err)
	}
	if util.IsFileNotExist(outFile) {
		t.Error("expected output file to be created")
	}
}

func TestGenCLATemplatePDFUnknownLang(t *testing.T) {
	pg := &pdfGenerator{
		pythonBin: "/bin/true",
		pdfOutDir: t.TempDir(),
		corp:      nil,
	}

	_, err := pg.GenCLATemplatePDF("link1", "/tmp/test.pdf", "chinese", nil)
	if err == nil {
		t.Fatal("expected error for unknown language")
	}
}

func TestGenCLATemplatePDFSuccess(t *testing.T) {
	eng, err := newGeneratorForEnglish()
	if err != nil {
		t.Fatalf("newGeneratorForEnglish failed: %v", err)
	}

	outDir := t.TempDir()
	pg := &pdfGenerator{
		pythonBin: "/bin/true",
		pdfOutDir: outDir,
		corp:      []*corpSigningPDF{eng},
	}

	claFile := filepath.Join(outDir, "dummy_cla.pdf")
	if err := os.WriteFile(claFile, []byte("%PDF-1.4 dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	outfile, err := pg.GenCLATemplatePDF("link1", claFile, "English", []models.CLAField{})
	if err != nil {
		t.Fatalf("GenCLATemplatePDF failed: %v", err)
	}
	if outfile == "" {
		t.Error("expected non-empty outfile path")
	}
}

func TestGenTemplateSignaturePDFExistingFile(t *testing.T) {
	corp, err := newGeneratorForEnglish()
	if err != nil {
		t.Fatalf("newGeneratorForEnglish failed: %v", err)
	}

	outFile := filepath.Join(t.TempDir(), "test_sig_existing.pdf")
	if err := os.WriteFile(outFile, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	err = genTemplateSignaturePDF(corp, []models.CLAField{}, outFile)
	if err != nil {
		t.Fatalf("genTemplateSignaturePDF with existing file failed: %v", err)
	}
	if util.IsFileNotExist(outFile) {
		t.Error("expected output file to be created")
	}
}

func TestGetPDFGenerator(t *testing.T) {
	if gen := GetPDFGenerator(); gen != nil {
		t.Error("expected nil generator before init")
	}

	InitPDFGenerator(&Config{
		PDFOutDir: t.TempDir(),
		PythonBin: "/bin/true",
	})

	if gen := GetPDFGenerator(); gen == nil {
		t.Error("expected non-nil generator after init")
	}
}
