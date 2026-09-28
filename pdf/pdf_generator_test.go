package pdf

import (
	"testing"

	"github.com/opensourceways/app-cla-server/models"
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
