package adapter

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestToFields(t *testing.T) {
	adapter := &linkAdatper{}

	fields := []domain.Field{
		{Id: "0", Required: true, CLAField: dp.CLAField{Type: "text", Title: "Name", Desc: "Your Name", MaxLength: 50}},
		{Id: "1", Required: false, CLAField: dp.CLAField{Type: "email", Title: "Email", Desc: "Your Email", MaxLength: 100}},
	}

	result := adapter.toFields(fields)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	if result[0].ID != "0" || result[0].Type != "text" || result[0].Title != "Name" {
		t.Errorf("unexpected result[0]: %+v", result[0])
	}
	if !result[0].Required {
		t.Error("expected Required=true for first field")
	}
	if result[1].Description != "Your Email" {
		t.Errorf("expected 'Your Email', got '%s'", result[1].Description)
	}
}

func TestToFieldsEmpty(t *testing.T) {
	adapter := &linkAdatper{}
	result := adapter.toFields([]domain.Field{})
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}

func TestLinkAdapterFields(t *testing.T) {
	_ = domain.Field{Id: "1", Required: true}
	_ = dp.CLAField{Type: "text", Title: "Name", Desc: "desc", MaxLength: 50}
}
