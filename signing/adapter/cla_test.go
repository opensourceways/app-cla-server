package adapter

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/app"
)

func TestIsAllowedPDFSource(t *testing.T) {
	adapter := &claAdatper{
		claPDFSource: []string{"https://gitee.com", "https://github.com"},
	}

	if !adapter.isAllowedPDFSource("https://gitee.com/repo/file.pdf") {
		t.Error("expected true for allowed source")
	}
	if !adapter.isAllowedPDFSource("https://github.com/repo/file.pdf") {
		t.Error("expected true for allowed source")
	}
	if adapter.isAllowedPDFSource("https://gitlab.com/repo/file.pdf") {
		t.Error("expected false for not allowed source")
	}
	if adapter.isAllowedPDFSource("http://gitee.com/repo/file.pdf") {
		t.Error("expected false for http (prefix mismatch)")
	}
}

func TestToCLADetail(t *testing.T) {
	adapter := &claAdatper{}
	dto := []app.CLADTO{
		{Id: "0", URL: "https://example.com/a.pdf", Language: "en"},
		{Id: "1", URL: "https://example.com/b.pdf", Language: "zh"},
	}

	result := adapter.toCLADetail(dto)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	if result[0].CLAId != "0" || result[0].URL != "https://example.com/a.pdf" || result[0].Language != "en" {
		t.Errorf("unexpected result[0]: %+v", result[0])
	}
	if result[1].CLAId != "1" || result[1].Language != "zh" {
		t.Errorf("unexpected result[1]: %+v", result[1])
	}
}

func TestToCLADetailEmpty(t *testing.T) {
	adapter := &claAdatper{}
	result := adapter.toCLADetail([]app.CLADTO{})
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}
