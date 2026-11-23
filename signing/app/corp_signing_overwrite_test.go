package app

import (
	"errors"
	"testing"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

type fakeCorpSigningRepoForOverwrite struct {
	fakeCorpSigningRepo

	findByEmailFn func(linkId string, email dp.EmailAddr) (domain.CorpSigning, bool, error)
	addFn         func(v *domain.CorpSigning) error
	updateAppFn   func(cs *domain.CorpSigning) error
	findByNameFn  func(linkId, corpName, corpDomain, excludeId string) (bool, error)

	addCalls        int
	updateAppCalls  int
	findByNameCalls int
	lastUpdated     *domain.CorpSigning
}

func (f *fakeCorpSigningRepoForOverwrite) FindByEmail(linkId string, email dp.EmailAddr) (domain.CorpSigning, bool, error) {
	if f.findByEmailFn != nil {
		return f.findByEmailFn(linkId, email)
	}
	return domain.CorpSigning{}, false, nil
}

func (f *fakeCorpSigningRepoForOverwrite) Add(v *domain.CorpSigning) error {
	f.addCalls++
	if f.addFn != nil {
		return f.addFn(v)
	}
	return nil
}

func (f *fakeCorpSigningRepoForOverwrite) UpdateApplication(cs *domain.CorpSigning) error {
	f.updateAppCalls++
	f.lastUpdated = cs
	if f.updateAppFn != nil {
		return f.updateAppFn(cs)
	}
	return nil
}

func (f *fakeCorpSigningRepoForOverwrite) FindByNameAndDomain(linkId, corpName, corpDomain, excludeId string) (bool, error) {
	f.findByNameCalls++
	if f.findByNameFn != nil {
		return f.findByNameFn(linkId, corpName, corpDomain, excludeId)
	}
	return false, nil
}

func newCorpSigningServiceForSignTest(
	repo repository.CorpSigning,
	vc vcservice.VCService,
) *corpSigningService {
	return &corpSigningService{
		repo: repo,
		vc:   verificationCodeService{vc: vc},
	}
}

func makeSignCmd(overwrite bool) *CmdToSignCorpCLA {
	return &CmdToSignCorpCLA{
		Link: domain.LinkInfo{
			Id:      "link1",
			CLAInfo: domain.CLAInfo{CLAId: "cla1", Language: dp.CreateLanguage("en")},
		},
		CorpName:         dp.CreateCorpName("TestCorp"),
		Rep:              domain.Representative{Name: dp.CreateName("Rep"), EmailAddr: dp.CreateEmailAddr("rep@test.com")},
		AllSingingInfo:   domain.AllSingingInfo{"f1": "v1"},
		VerificationCode: "123456",
		Overwrite:        overwrite,
	}
}

func pendingCorpSigning() domain.CorpSigning {
	return domain.CorpSigning{
		Id:      "cs1",
		Date:    "2024-01-01",
		Version: 1,
		Link: domain.LinkInfo{
			Id:      "link1",
			CLAInfo: domain.CLAInfo{CLAId: "cla-old", Language: dp.CreateLanguage("zh")},
		},
		Rep: domain.Representative{
			Name:      dp.CreateName("OldRep"),
			EmailAddr: dp.CreateEmailAddr("rep@test.com"),
		},
		Corp:    domain.NewCorporation(dp.CreateCorpName("OldCorp"), dp.CreateEmailAddr("rep@test.com")),
		AllInfo: domain.AllSingingInfo{"f1": "old"},
	}
}

func approvedCorpSigning() domain.CorpSigning {
	cs := pendingCorpSigning()
	cs.Admin = domain.Manager{
		Id: "admin1",
		Representative: domain.Representative{
			Name:      dp.CreateName("OldRep"),
			EmailAddr: dp.CreateEmailAddr("rep@test.com"),
		},
	}
	return cs
}

func TestSignNoHistoryCreatesNew(t *testing.T) {
	repo := &fakeCorpSigningRepoForOverwrite{
		fakeCorpSigningRepo: fakeCorpSigningRepo{},
		findByEmailFn: func(linkId string, email dp.EmailAddr) (domain.CorpSigning, bool, error) {
			return domain.CorpSigning{}, false, nil
		},
	}
	s := newCorpSigningServiceForSignTest(repo, &fakeVCService{})

	if err := s.Sign(makeSignCmd(false)); err != nil {
		t.Fatalf("Sign: unexpected error %v", err)
	}
	if repo.addCalls != 1 {
		t.Errorf("Add calls: got %d, want 1", repo.addCalls)
	}
	if repo.updateAppCalls != 0 {
		t.Errorf("UpdateApplication should not be called, got %d", repo.updateAppCalls)
	}
}

func TestSignPendingNoOverwriteReturnsPendingExists(t *testing.T) {
	repo := &fakeCorpSigningRepoForOverwrite{
		fakeCorpSigningRepo: fakeCorpSigningRepo{},
		findByEmailFn: func(linkId string, email dp.EmailAddr) (domain.CorpSigning, bool, error) {
			return pendingCorpSigning(), true, nil
		},
	}
	s := newCorpSigningServiceForSignTest(repo, &fakeVCService{})

	err := s.Sign(makeSignCmd(false))
	if err == nil {
		t.Fatal("Sign: expected error for pending without overwrite")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeCorpSigningPendingExists) {
		t.Errorf("expected error code %q, got %v", domain.ErrorCodeCorpSigningPendingExists, err)
	}
	if repo.addCalls != 0 {
		t.Errorf("Add should not be called, got %d", repo.addCalls)
	}
	if repo.updateAppCalls != 0 {
		t.Errorf("UpdateApplication should not be called, got %d", repo.updateAppCalls)
	}
}

func TestSignPendingWithOverwriteUpdatesApplication(t *testing.T) {
	repo := &fakeCorpSigningRepoForOverwrite{
		fakeCorpSigningRepo: fakeCorpSigningRepo{},
		findByEmailFn: func(linkId string, email dp.EmailAddr) (domain.CorpSigning, bool, error) {
			return pendingCorpSigning(), true, nil
		},
	}
	s := newCorpSigningServiceForSignTest(repo, &fakeVCService{})

	if err := s.Sign(makeSignCmd(true)); err != nil {
		t.Fatalf("Sign: unexpected error %v", err)
	}
	if repo.updateAppCalls != 1 {
		t.Errorf("UpdateApplication calls: got %d, want 1", repo.updateAppCalls)
	}
	if repo.addCalls != 0 {
		t.Errorf("Add should not be called, got %d", repo.addCalls)
	}
	if repo.lastUpdated != nil {
		if repo.lastUpdated.Corp.Name.CorpName() != "TestCorp" {
			t.Errorf("updated corp name: got %s, want TestCorp", repo.lastUpdated.Corp.Name.CorpName())
		}
		if repo.lastUpdated.Rep.Name.Name() != "Rep" {
			t.Errorf("updated rep name: got %s, want Rep", repo.lastUpdated.Rep.Name.Name())
		}
		if repo.lastUpdated.Link.CLAId != "cla1" {
			t.Errorf("updated cla_id: got %s, want cla1", repo.lastUpdated.Link.CLAId)
		}
		if repo.lastUpdated.HasPDF {
			t.Error("HasPDF should be false after overwrite")
		}
	}
}

func TestSignAlreadyApprovedReturnsAlreadyApproved(t *testing.T) {
	repo := &fakeCorpSigningRepoForOverwrite{
		fakeCorpSigningRepo: fakeCorpSigningRepo{},
		findByEmailFn: func(linkId string, email dp.EmailAddr) (domain.CorpSigning, bool, error) {
			return approvedCorpSigning(), true, nil
		},
	}
	s := newCorpSigningServiceForSignTest(repo, &fakeVCService{})

	err := s.Sign(makeSignCmd(true))
	if err == nil {
		t.Fatal("Sign: expected error for already approved")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeCorpSigningAlreadyApproved) {
		t.Errorf("expected error code %q, got %v", domain.ErrorCodeCorpSigningAlreadyApproved, err)
	}
	if repo.addCalls != 0 {
		t.Errorf("Add should not be called, got %d", repo.addCalls)
	}
	if repo.updateAppCalls != 0 {
		t.Errorf("UpdateApplication should not be called, got %d", repo.updateAppCalls)
	}
}

func TestSignOverwriteWithCorpNameConflictReturnsReSigning(t *testing.T) {
	repo := &fakeCorpSigningRepoForOverwrite{
		fakeCorpSigningRepo: fakeCorpSigningRepo{},
		findByEmailFn: func(linkId string, email dp.EmailAddr) (domain.CorpSigning, bool, error) {
			return pendingCorpSigning(), true, nil
		},
		findByNameFn: func(linkId, corpName, corpDomain, excludeId string) (bool, error) {
			return true, nil
		},
	}
	s := newCorpSigningServiceForSignTest(repo, &fakeVCService{})

	err := s.Sign(makeSignCmd(true))
	if err == nil {
		t.Fatal("Sign: expected error for corp name conflict")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeCorpSigningReSigning) {
		t.Errorf("expected error code %q, got %v", domain.ErrorCodeCorpSigningReSigning, err)
	}
	if repo.updateAppCalls != 0 {
		t.Errorf("UpdateApplication should not be called on conflict, got %d", repo.updateAppCalls)
	}
}

func TestSignFindByEmailErrorPropagates(t *testing.T) {
	repo := &fakeCorpSigningRepoForOverwrite{
		fakeCorpSigningRepo: fakeCorpSigningRepo{},
		findByEmailFn: func(linkId string, email dp.EmailAddr) (domain.CorpSigning, bool, error) {
			return domain.CorpSigning{}, false, errors.New("db error")
		},
	}
	s := newCorpSigningServiceForSignTest(repo, &fakeVCService{})

	if err := s.Sign(makeSignCmd(false)); err == nil {
		t.Error("Sign should propagate FindByEmail error")
	}
}

func TestSignOverwriteUpdateApplicationErrorPropagates(t *testing.T) {
	repo := &fakeCorpSigningRepoForOverwrite{
		fakeCorpSigningRepo: fakeCorpSigningRepo{},
		findByEmailFn: func(linkId string, email dp.EmailAddr) (domain.CorpSigning, bool, error) {
			return pendingCorpSigning(), true, nil
		},
		updateAppFn: func(cs *domain.CorpSigning) error {
			return commonRepo.NewErrorConcurrentUpdating(errors.New("version mismatch"))
		},
	}
	s := newCorpSigningServiceForSignTest(repo, &fakeVCService{})

	if err := s.Sign(makeSignCmd(true)); err == nil {
		t.Error("Sign should propagate UpdateApplication error")
	}
}

func TestSignNoHistoryAddDuplicateReturnsReSigning(t *testing.T) {
	repo := &fakeCorpSigningRepoForOverwrite{
		fakeCorpSigningRepo: fakeCorpSigningRepo{},
		findByEmailFn: func(linkId string, email dp.EmailAddr) (domain.CorpSigning, bool, error) {
			return domain.CorpSigning{}, false, nil
		},
		addFn: func(v *domain.CorpSigning) error {
			return commonRepo.NewErrorDuplicateCreating(errors.New("duplicate"))
		},
	}
	s := newCorpSigningServiceForSignTest(repo, &fakeVCService{})

	err := s.Sign(makeSignCmd(false))
	if err == nil {
		t.Fatal("Sign: expected error for duplicate Add")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeCorpSigningReSigning) {
		t.Errorf("expected error code %q, got %v", domain.ErrorCodeCorpSigningReSigning, err)
	}
}

func TestSignVerificationCodeErrorPropagates(t *testing.T) {
	repo := &fakeCorpSigningRepoForOverwrite{
		fakeCorpSigningRepo: fakeCorpSigningRepo{},
	}
	s := newCorpSigningServiceForSignTest(repo, &fakeVCService{verifyErr: errors.New("bad code")})

	if err := s.Sign(makeSignCmd(false)); err == nil {
		t.Error("Sign should propagate verification code error")
	}
}

func TestSignOverwriteNoConflictChecksNameThenUpdates(t *testing.T) {
	var nameCheckArgs struct {
		linkId, corpName, corpDomain, excludeId string
	}
	repo := &fakeCorpSigningRepoForOverwrite{
		fakeCorpSigningRepo: fakeCorpSigningRepo{},
		findByEmailFn: func(linkId string, email dp.EmailAddr) (domain.CorpSigning, bool, error) {
			return pendingCorpSigning(), true, nil
		},
		findByNameFn: func(linkId, corpName, corpDomain, excludeId string) (bool, error) {
			nameCheckArgs.linkId = linkId
			nameCheckArgs.corpName = corpName
			nameCheckArgs.corpDomain = corpDomain
			nameCheckArgs.excludeId = excludeId
			return false, nil
		},
	}
	s := newCorpSigningServiceForSignTest(repo, &fakeVCService{})

	if err := s.Sign(makeSignCmd(true)); err != nil {
		t.Fatalf("Sign: unexpected error %v", err)
	}
	if nameCheckArgs.linkId != "link1" {
		t.Errorf("name check linkId: got %s, want link1", nameCheckArgs.linkId)
	}
	if nameCheckArgs.corpName != "TestCorp" {
		t.Errorf("name check corpName: got %s, want TestCorp", nameCheckArgs.corpName)
	}
	if nameCheckArgs.corpDomain != "test.com" {
		t.Errorf("name check corpDomain: got %s, want test.com", nameCheckArgs.corpDomain)
	}
	if nameCheckArgs.excludeId != "cs1" {
		t.Errorf("name check excludeId: got %s, want cs1", nameCheckArgs.excludeId)
	}
}
