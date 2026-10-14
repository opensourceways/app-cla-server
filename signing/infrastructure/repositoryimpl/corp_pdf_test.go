package repositoryimpl

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestCorpSigningDOPDFURLField(t *testing.T) {
	do := corpSigningDO{
		PDFURL: "cla-pdfs/corp-signing/abc.pdf",
		HasPDF: true,
	}

	if do.PDFURL != "cla-pdfs/corp-signing/abc.pdf" {
		t.Errorf("PDFURL: got %q, want %q", do.PDFURL, "cla-pdfs/corp-signing/abc.pdf")
	}
	if !do.HasPDF {
		t.Error("HasPDF should be true")
	}
}

func TestCorpSigningDOPDFFieldRetained(t *testing.T) {
	pdfData := []byte("legacy-pdf-bytes")
	do := corpSigningDO{
		PDF:    pdfData,
		HasPDF: true,
	}

	if string(do.PDF) != "legacy-pdf-bytes" {
		t.Errorf("PDF: got %q, want %q", string(do.PDF), "legacy-pdf-bytes")
	}
}

func TestCorpSigningDOBothFieldsCoexist(t *testing.T) {
	do := corpSigningDO{
		PDFURL: "obs://bucket/key.pdf",
		PDF:    []byte("legacy"),
		HasPDF: true,
	}

	if do.PDFURL != "obs://bucket/key.pdf" {
		t.Errorf("PDFURL: got %q", do.PDFURL)
	}
	if string(do.PDF) != "legacy" {
		t.Errorf("PDF: got %q", string(do.PDF))
	}
}

func TestCorpSigningDOPDFURLBsonTag(t *testing.T) {
	cs := &domain.CorpSigning{
		Id:   "abc",
		Date: "2026-01-01",
		Link: domain.LinkInfo{Id: "link1", CLAInfo: domain.CLAInfo{CLAId: "cla1", Language: dp.CreateLanguage("en")}},
		Rep:  domain.Representative{Name: dp.CreateName("Rep"), EmailAddr: dp.CreateEmailAddr("rep@test.com")},
		Corp: domain.Corporation{Name: dp.CreateCorpName("Corp"), PrimaryEmailDomain: "test.com", AllEmailDomains: []string{"test.com"}},
	}

	do := toCorpSigningDO(cs)

	if do.PDFURL != "" {
		t.Errorf("toCorpSigningDO: PDFURL should be empty for new corp signing, got %q", do.PDFURL)
	}
	if do.PDF != nil {
		t.Errorf("toCorpSigningDO: PDF should be nil for new corp signing")
	}
}
