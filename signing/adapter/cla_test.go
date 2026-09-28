package adapter

import (
	"testing"
	"time"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestFormatCLADate(t *testing.T) {
	tests := []struct {
		name string
		unix int64
		want string
	}{
		{"zero (历史遗留)", 0, "1970-01-01"},
		{"negative", -1, "1970-01-01"},
		{"valid timestamp", 1726099200, "2024-09-12"},
		{"valid timestamp 2", 1704067200, "2024-01-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCLADate(tt.unix)
			if got != tt.want {
				t.Errorf("formatCLADate(%d) = %s, want %s", tt.unix, got, tt.want)
			}
		})
	}
}

func TestToCLADetailUpdatedAt(t *testing.T) {
	adapter := &claAdatper{}

	ts := time.Now().Unix()
	dto := []app.CLADTO{
		{Id: "0", URL: "http://example.com/a.pdf", Language: "zh", UpdatedAt: ts},
		{Id: "1", URL: "http://example.com/b.pdf", Language: "en", UpdatedAt: 0},
	}

	result := adapter.toCLADetail(dto)

	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}

	// first item: valid timestamp
	wantFirst := formatCLADate(ts)
	if result[0].UpdatedAt != wantFirst {
		t.Errorf("result[0].UpdatedAt = %s, want %s", result[0].UpdatedAt, wantFirst)
	}
	if result[0].CLAId != "0" {
		t.Errorf("result[0].CLAId = %s, want 0", result[0].CLAId)
	}

	// second item: zero timestamp -> 1970-01-01
	if result[1].UpdatedAt != "1970-01-01" {
		t.Errorf("result[1].UpdatedAt = %s, want 1970-01-01", result[1].UpdatedAt)
	}
}

func TestToCLADetailEmpty(t *testing.T) {
	adapter := &claAdatper{}
	result := adapter.toCLADetail([]app.CLADTO{})

	if len(result) != 0 {
		t.Errorf("expected 0 items, got %d", len(result))
	}
}

func TestToCLAFields(t *testing.T) {
	fields := []domain.Field{
		{
			Id:       "1",
			Required: true,
			CLAField: dp.CLAField{
				Type:  "name",
				Desc:  "description1",
				Title: "Name",
			},
		},
		{
			Id:       "2",
			Required: false,
			CLAField: dp.CLAField{
				Type:  "email",
				Desc:  "description2",
				Title: "Email",
			},
		},
	}

	result := toCLAFields(fields)

	if len(result) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(result))
	}

	if result[0].ID != "1" || result[0].Title != "Name" || result[0].Type != "name" {
		t.Errorf("result[0] = %+v", result[0])
	}
	if !result[0].Required {
		t.Error("result[0] should be required")
	}

	if result[1].ID != "2" || result[1].Title != "Email" || result[1].Type != "email" {
		t.Errorf("result[1] = %+v", result[1])
	}
	if result[1].Required {
		t.Error("result[1] should not be required")
	}
}

func TestToCLAFieldsEmpty(t *testing.T) {
	result := toCLAFields([]domain.Field{})

	if len(result) != 0 {
		t.Errorf("expected 0 fields, got %d", len(result))
	}
}

func TestDiffPreviewNotAllowedSource(t *testing.T) {
	adapter := &claAdatper{
		claPDFSource: []string{"https://gitee.com"},
	}

	opt := &models.CLADiffPreviewOpt{
		URL:      "https://evil.com/x.pdf",
		Type:     "corporation",
		Language: "zh",
	}

	_, merr := adapter.DiffPreview("user1", "link1", opt)
	if merr == nil {
		t.Fatal("expected error for not allowed source")
	}
	if !merr.IsErrorOf(models.ErrNotAllowedCLAPDFSource) {
		t.Errorf("expected ErrNotAllowedCLAPDFSource, got %s", merr.ErrCode())
	}
}

func TestDiffPreviewInvalidType(t *testing.T) {
	adapter := &claAdatper{
		claPDFSource:         []string{"https://gitee.com"},
		fileTypeOfCLAContent: "pdf",
		maxSizeOfCLAContent:  2 << 20,
	}

	opt := &models.CLADiffPreviewOpt{
		URL:      "https://gitee.com/x.pdf",
		Type:     "invalid_type",
		Language: "zh",
	}

	_, merr := adapter.DiffPreview("user1", "link1", opt)
	if merr == nil {
		t.Fatal("expected error for invalid type")
	}
	if !merr.IsErrorOf(models.ErrBadRequestParameter) {
		t.Errorf("expected ErrBadRequestParameter, got %s", merr.ErrCode())
	}
}
