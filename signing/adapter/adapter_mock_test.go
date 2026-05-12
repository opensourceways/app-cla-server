package adapter

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

// Light mock that only implements what we need for constructor/simple tests
type mockCLASvc struct{}

func (m *mockCLASvc) Add(cmd *app.CmdToAddCLA) error                                 { return nil }
func (m *mockCLASvc) Update(cmd *app.CmdToUpdateCLA) error                            { return nil }
func (m *mockCLASvc) Remove(cmd *app.CmdToRemoveCLA) error                            { return nil }
func (m *mockCLASvc) List(userId, linkId string) ([]app.CLADTO, []app.CLADTO, error) { return nil, nil, nil }
func (m *mockCLASvc) CLALocalFilePath(index domain.CLAIndex) string                   { return "" }

func TestCLAAdapterLocalPath(t *testing.T) {
	adapter := &claAdatper{
		s: &mockCLASvc{},
	}
	// Just verifying the adapter struct works
	_ = adapter
}

func TestCLAAdapterStruct(t *testing.T) {
	adapter := NewCLAAdapter(&mockCLASvc{}, 100, "pdf", []string{"https://gitee.com"})
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
	if adapter.maxSizeOfCLAContent != 100 {
		t.Errorf("expected 100, got %d", adapter.maxSizeOfCLAContent)
	}
	if adapter.fileTypeOfCLAContent != "pdf" {
		t.Errorf("expected 'pdf', got '%s'", adapter.fileTypeOfCLAContent)
	}
}

func TestNewLinkAdapter(t *testing.T) {
	adapter := NewLinkAdapter(nil, nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewUserAdapter(t *testing.T) {
	adapter := NewUserAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewAccessTokenAdapter(t *testing.T) {
	adapter := NewAccessTokenAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewMigrationAdapter(t *testing.T) {
	adapter := NewMigrationAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewSMTPAdapter2(t *testing.T) {
	adapter := NewSMTPAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewCorpSigningAdapter2(t *testing.T) {
	adapter := NewCorpSigningAdapter(nil, nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewIndividualSigningAdapter2(t *testing.T) {
	adapter := NewIndividualSigningAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewEmployeeSigningAdapter2(t *testing.T) {
	adapter := NewEmployeeSigningAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewEmployeeManagerAdapter2(t *testing.T) {
	adapter := NewEmployeeManagerAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewCorpEmailDomainAdapter2(t *testing.T) {
	adapter := NewCorpEmailDomainAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewCorpAdminAdapter2(t *testing.T) {
	adapter := NewCorpAdminAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewCorpPDFAdapter2(t *testing.T) {
	adapter := NewCorpPDFAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNewSMTPAdapter(t *testing.T) {
	adapter := NewSMTPAdapter(nil)
	if adapter == nil {
		t.Fatal("expected non-nil")
	}
}
