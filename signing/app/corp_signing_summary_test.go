package app

import (
	"errors"
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

type fakeCorpSigningRepoForSummary struct {
	fakeCorpSigningRepo

	findCorpSummaryFn func(linkId, domain string) ([]repository.CorpSummary, error)

	findCorpSummaryCalledWith struct {
		linkId string
		domain string
	}
}

func (f *fakeCorpSigningRepoForSummary) FindCorpSummary(linkId, domain string) ([]repository.CorpSummary, error) {
	f.findCorpSummaryCalledWith.linkId = linkId
	f.findCorpSummaryCalledWith.domain = domain
	if f.findCorpSummaryFn != nil {
		return f.findCorpSummaryFn(linkId, domain)
	}
	return nil, nil
}

// ---- FindCorpSummary ----

func TestFindCorpSummaryOnlyReturnsCorpsWithAdmin(t *testing.T) {
	repo := &fakeCorpSigningRepoForSummary{
		findCorpSummaryFn: func(linkId, domain string) ([]repository.CorpSummary, error) {
			return []repository.CorpSummary{
				{CorpName: dp.CreateCorpName("approved corp"), HasAdmin: true, CorpSigningId: "cs1"},
				{CorpName: dp.CreateCorpName("pending corp"), CorpSigningId: "cs2"},
				{CorpName: dp.CreateCorpName("approved corp 2"), HasAdmin: true, CorpSigningId: "cs3"},
			}, nil
		},
	}

	s := &corpSigningService{repo: repo}

	cmd := &CmdToFindCorpSummary{
		LinkId:    "link1",
		EmailAddr: dp.CreateEmailAddr("dev@example.com"),
	}

	r, err := s.FindCorpSummary(cmd)
	if err != nil {
		t.Fatalf("FindCorpSummary: unexpected error %v", err)
	}

	if len(r) != 2 {
		t.Fatalf("result count: got %d, want 2", len(r))
	}
	if r[0].CorpName != "approved corp" || r[0].CorpSigningId != "cs1" {
		t.Errorf("r[0]: got %s/%s, want approved corp/cs1", r[0].CorpName, r[0].CorpSigningId)
	}
	if r[1].CorpName != "approved corp 2" || r[1].CorpSigningId != "cs3" {
		t.Errorf("r[1]: got %s/%s, want approved corp 2/cs3", r[1].CorpName, r[1].CorpSigningId)
	}

	if repo.findCorpSummaryCalledWith.linkId != "link1" {
		t.Errorf("linkId: got %s, want link1", repo.findCorpSummaryCalledWith.linkId)
	}
	if repo.findCorpSummaryCalledWith.domain != "example.com" {
		t.Errorf("domain: got %s, want example.com", repo.findCorpSummaryCalledWith.domain)
	}
}

func TestFindCorpSummaryReturnsCorpWithAdminButNoManagers(t *testing.T) {
	// Regression: a corp that has a super admin email but no employee managers
	// must still be returned, so employees can sign (the admin can approve).
	repo := &fakeCorpSigningRepoForSummary{
		findCorpSummaryFn: func(linkId, domain string) ([]repository.CorpSummary, error) {
			return []repository.CorpSummary{
				{CorpName: dp.CreateCorpName("admin only corp"), HasAdmin: true, CorpSigningId: "cs1"},
			}, nil
		},
	}

	s := &corpSigningService{repo: repo}

	cmd := &CmdToFindCorpSummary{
		LinkId:    "link1",
		EmailAddr: dp.CreateEmailAddr("dev@example.com"),
	}

	r, err := s.FindCorpSummary(cmd)
	if err != nil {
		t.Fatalf("FindCorpSummary: unexpected error %v", err)
	}
	if len(r) != 1 {
		t.Fatalf("result count: got %d, want 1 (admin-only corp must be returned)", len(r))
	}
	if r[0].CorpName != "admin only corp" || r[0].CorpSigningId != "cs1" {
		t.Errorf("r[0]: got %s/%s, want admin only corp/cs1", r[0].CorpName, r[0].CorpSigningId)
	}
}

func TestFindCorpSummaryAllWithoutAdminReturnsEmpty(t *testing.T) {
	repo := &fakeCorpSigningRepoForSummary{
		findCorpSummaryFn: func(linkId, domain string) ([]repository.CorpSummary, error) {
			return []repository.CorpSummary{
				{CorpName: dp.CreateCorpName("pending corp"), CorpSigningId: "cs1"},
			}, nil
		},
	}

	s := &corpSigningService{repo: repo}

	cmd := &CmdToFindCorpSummary{
		LinkId:    "link1",
		EmailAddr: dp.CreateEmailAddr("dev@example.com"),
	}

	r, err := s.FindCorpSummary(cmd)
	if err != nil {
		t.Fatalf("FindCorpSummary: unexpected error %v", err)
	}
	if len(r) != 0 {
		t.Errorf("result count: got %d, want 0", len(r))
	}
}

func TestFindCorpSummaryRepoError(t *testing.T) {
	repo := &fakeCorpSigningRepoForSummary{
		findCorpSummaryFn: func(linkId, domain string) ([]repository.CorpSummary, error) {
			return nil, errors.New("db error")
		},
	}

	s := &corpSigningService{repo: repo}

	cmd := &CmdToFindCorpSummary{
		LinkId:    "link1",
		EmailAddr: dp.CreateEmailAddr("dev@example.com"),
	}

	if _, err := s.FindCorpSummary(cmd); err == nil {
		t.Error("FindCorpSummary should propagate repo error")
	}
}
