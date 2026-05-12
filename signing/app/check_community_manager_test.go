package app

import (
	"testing"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

type mockLinkRepo struct {
	findRes domain.Link
	findErr error
}

func (m *mockLinkRepo) Find(linkId string) (domain.Link, error) { return m.findRes, m.findErr }
func (m *mockLinkRepo) FindAll(string) ([]repository.LinkSummary, error) { return nil, nil }
func (m *mockLinkRepo) NewLinkId() string                               { return "" }
func (m *mockLinkRepo) Add(*domain.Link) error                          { return nil }
func (m *mockLinkRepo) Remove(*domain.Link) error                       { return nil }
func (m *mockLinkRepo) AddCLA(*domain.Link, *domain.CLA) error          { return nil }
func (m *mockLinkRepo) UpdateCLA(*domain.Link, *domain.CLA) error       { return nil }
func (m *mockLinkRepo) RemoveCLA(*domain.Link, *domain.CLA) error       { return nil }
func (m *mockLinkRepo) ListAll() ([]repository.LinkCLA, error)          { return nil, nil }

func TestCheckIfCommunityManagerSuccess(t *testing.T) {
	repo := &mockLinkRepo{
		findRes: domain.Link{
			Id:        "link-1",
			Submitter: "admin",
		},
	}
	link, err := checkIfCommunityManager("admin", "link-1", repo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if link.Id != "link-1" {
		t.Errorf("expected 'link-1', got '%s'", link.Id)
	}
}

func TestCheckIfCommunityManagerNotFound(t *testing.T) {
	repo := &mockLinkRepo{
		findErr: commonRepo.NewErrorResourceNotFound(nil),
	}
	_, err := checkIfCommunityManager("admin", "link-1", repo)
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestCheckIfCommunityManagerNoPermission(t *testing.T) {
	repo := &mockLinkRepo{
		findRes: domain.Link{
			Id:        "link-1",
			Submitter: "admin",
		},
	}
	_, err := checkIfCommunityManager("other-user", "link-1", repo)
	if err == nil {
		t.Error("expected error for no permission")
	}
}
