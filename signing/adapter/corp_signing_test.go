package adapter

import (
	"testing"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/app"
)

// mockCorpSigningService embeds the interface so only the tested methods
// need overriding. Calling an un-overridden method panics, which is
// acceptable in unit tests that only exercise specific paths.
type mockCorpSigningService struct {
	app.CorpSigningService
	listPageFn func(userId, linkId string, page, pageSize int, adminAdded bool, searchQuery string) (app.CorpSigningPageDTO, error)
	listFn     func(userId, linkId string) ([]app.CorpSigningDTO, error)
}

func (m *mockCorpSigningService) ListPage(userId, linkId string, page, pageSize int, adminAdded bool, searchQuery string) (app.CorpSigningPageDTO, error) {
	return m.listPageFn(userId, linkId, page, pageSize, adminAdded, searchQuery)
}

func (m *mockCorpSigningService) List(userId, linkId string) ([]app.CorpSigningDTO, error) {
	return m.listFn(userId, linkId)
}

func TestCorpSigningAdapterListPageAdminAddedDate(t *testing.T) {
	adapter := NewCorpSigningAdapter(&mockCorpSigningService{
		listPageFn: func(userId, linkId string, page, pageSize int, adminAdded bool, searchQuery string) (app.CorpSigningPageDTO, error) {
			return app.CorpSigningPageDTO{
				Total: 1,
				Data: []app.CorpSigningDTO{
					{
						Id:             "cs1",
						Date:           "2026-08-30",
						Language:       "zh",
						CorpName:       "华为技术有限公司",
						AdminAddedDate: "2026-08-31",
					},
				},
			}, nil
		},
	}, nil)

	result, merr := adapter.ListPage("user1", "link1", 1, 10, true, "华为")
	if merr != nil {
		t.Fatalf("ListPage error: %v", merr)
	}
	if result.Total != 1 {
		t.Fatalf("Total = %d, want 1", result.Total)
	}
	if len(result.Data) != 1 {
		t.Fatalf("Data len = %d, want 1", len(result.Data))
	}
	if result.Data[0].AdminAddedDate != "2026-08-31" {
		t.Errorf("AdminAddedDate = %q, want 2026-08-31", result.Data[0].AdminAddedDate)
	}
	if result.Data[0].CorporationName != "华为技术有限公司" {
		t.Errorf("CorporationName = %q, want 华为技术有限公司", result.Data[0].CorporationName)
	}
}

func TestCorpSigningAdapterListAdminAddedDate(t *testing.T) {
	adapter := NewCorpSigningAdapter(&mockCorpSigningService{
		listFn: func(userId, linkId string) ([]app.CorpSigningDTO, error) {
			return []app.CorpSigningDTO{
				{
					Id:             "cs1",
					Date:           "2026-08-30",
					AdminAddedDate: "2026-08-31",
				},
				{
					Id:             "cs2",
					Date:           "2026-09-01",
					AdminAddedDate: "",
				},
			}, nil
		},
	}, nil)

	result, merr := adapter.List("user1", "link1")
	if merr != nil {
		t.Fatalf("List error: %v", merr)
	}
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	if result[0].AdminAddedDate != "2026-08-31" {
		t.Errorf("result[0] AdminAddedDate = %q, want 2026-08-31", result[0].AdminAddedDate)
	}
	if result[1].AdminAddedDate != "" {
		t.Errorf("result[1] AdminAddedDate = %q, want empty", result[1].AdminAddedDate)
	}
}

func TestCorpSigningAdapterListPageInvalidParam(t *testing.T) {
	adapter := NewCorpSigningAdapter(&mockCorpSigningService{}, nil)

	_, merr := adapter.ListPage("user1", "link1", 0, 10, true, "")
	if merr == nil {
		t.Error("expected error for page=0")
	}

	_, merr = adapter.ListPage("user1", "link1", 1, 0, true, "")
	if merr == nil {
		t.Error("expected error for pageSize=0")
	}
}

// Ensure models import is used.
var _ = models.CorporationSigningSummary{}
