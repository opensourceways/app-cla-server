package app

import (
	"errors"
	"testing"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

// fakeCorpSigningRepoForEM extends fakeCorpSigningRepo with AddEmployeeManagers support.
type fakeCorpSigningRepoForEM struct {
	fakeCorpSigningRepo

	addEmployeeManagersFn func(cs *domain.CorpSigning, ms []domain.Manager) error

	addEmployeeManagersCalls int
	lastManagers             []domain.Manager
}

func (f *fakeCorpSigningRepoForEM) AddEmployeeManagers(cs *domain.CorpSigning, ms []domain.Manager) error {
	f.addEmployeeManagersCalls++
	f.lastManagers = ms
	if f.addEmployeeManagersFn != nil {
		return f.addEmployeeManagersFn(cs, ms)
	}
	return nil
}

func newEmployeeManagerServiceForTest(repo repository.CorpSigning) *employeeManagerService {
	return &employeeManagerService{repo: repo}
}

func TestSelfAssignManagerSuccess(t *testing.T) {
	adminEmail := dp.CreateEmailAddr("admin@test.com")
	cs := domain.CorpSigning{
		Id: "cs1",
		Admin: domain.Manager{
			Id: "admin1",
			Representative: domain.Representative{
				Name:      dp.CreateName("Admin"),
				EmailAddr: adminEmail,
			},
		},
		Corp: domain.Corporation{AllEmailDomains: []string{"test.com"}},
	}

	repo := &fakeCorpSigningRepoForEM{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return cs, nil
			},
		},
	}

	s := newEmployeeManagerServiceForTest(repo)

	if err := s.SelfAssignManager("cs1"); err != nil {
		t.Fatalf("SelfAssignManager: unexpected error %v", err)
	}

	if repo.addEmployeeManagersCalls != 1 {
		t.Errorf("AddEmployeeManagers calls: got %d, want 1", repo.addEmployeeManagersCalls)
	}

	if len(repo.lastManagers) != 1 {
		t.Fatalf("lastManagers length: got %d, want 1", len(repo.lastManagers))
	}

	if repo.lastManagers[0].Id != "self" {
		t.Errorf("lastManagers[0].Id: got %s, want self", repo.lastManagers[0].Id)
	}
}

func TestSelfAssignManagerAdminNotExists(t *testing.T) {
	cs := domain.CorpSigning{
		Id:   "cs1",
		Corp: domain.Corporation{AllEmailDomains: []string{"test.com"}},
	}

	repo := &fakeCorpSigningRepoForEM{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return cs, nil
			},
		},
	}

	s := newEmployeeManagerServiceForTest(repo)

	err := s.SelfAssignManager("cs1")
	if err == nil {
		t.Fatal("SelfAssignManager should fail when admin is empty")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeCorpAdminNotExists) {
		t.Errorf("expected error code %q, got %v", domain.ErrorCodeCorpAdminNotExists, err)
	}
	if repo.addEmployeeManagersCalls != 0 {
		t.Errorf("AddEmployeeManagers should not be called, got %d calls", repo.addEmployeeManagersCalls)
	}
}

func TestSelfAssignManagerFindError(t *testing.T) {
	repo := &fakeCorpSigningRepoForEM{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return domain.CorpSigning{}, errors.New("db error")
			},
		},
	}

	s := newEmployeeManagerServiceForTest(repo)

	if err := s.SelfAssignManager("cs1"); err == nil {
		t.Error("SelfAssignManager should propagate Find error")
	}
	if repo.addEmployeeManagersCalls != 0 {
		t.Errorf("AddEmployeeManagers should not be called when Find fails, got %d calls", repo.addEmployeeManagersCalls)
	}
}

func TestSelfAssignManagerNotFound(t *testing.T) {
	repo := &fakeCorpSigningRepoForEM{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return domain.CorpSigning{}, commonRepo.NewErrorResourceNotFound(errors.New("not found"))
			},
		},
	}

	s := newEmployeeManagerServiceForTest(repo)

	err := s.SelfAssignManager("cs1")
	if err == nil {
		t.Fatal("SelfAssignManager should fail for missing corp signing")
	}
	if repo.addEmployeeManagersCalls != 0 {
		t.Errorf("AddEmployeeManagers should not be called when Find fails, got %d calls", repo.addEmployeeManagersCalls)
	}
}

func TestSelfAssignManagerAddEmployeeManagersError(t *testing.T) {
	adminEmail := dp.CreateEmailAddr("admin@test.com")
	cs := domain.CorpSigning{
		Id: "cs1",
		Admin: domain.Manager{
			Id: "admin1",
			Representative: domain.Representative{
				Name:      dp.CreateName("Admin"),
				EmailAddr: adminEmail,
			},
		},
		Corp: domain.Corporation{AllEmailDomains: []string{"test.com"}},
	}

	repo := &fakeCorpSigningRepoForEM{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return cs, nil
			},
		},
		addEmployeeManagersFn: func(cs *domain.CorpSigning, ms []domain.Manager) error {
			return errors.New("db write error")
		},
	}

	s := newEmployeeManagerServiceForTest(repo)

	if err := s.SelfAssignManager("cs1"); err == nil {
		t.Error("SelfAssignManager should propagate AddEmployeeManagers error")
	}
	if repo.addEmployeeManagersCalls != 1 {
		t.Errorf("AddEmployeeManagers calls: got %d, want 1", repo.addEmployeeManagersCalls)
	}
}

func TestSelfAssignManagerDuplicate(t *testing.T) {
	adminEmail := dp.CreateEmailAddr("admin@test.com")
	cs := domain.CorpSigning{
		Id: "cs1",
		Admin: domain.Manager{
			Id: "admin1",
			Representative: domain.Representative{
				Name:      dp.CreateName("Admin"),
				EmailAddr: adminEmail,
			},
		},
		Corp: domain.Corporation{AllEmailDomains: []string{"test.com"}},
		Managers: []domain.Manager{
			{
				Id: "self",
				Representative: domain.Representative{
					Name:      dp.CreateName("Admin"),
					EmailAddr: adminEmail,
				},
			},
		},
	}

	repo := &fakeCorpSigningRepoForEM{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return cs, nil
			},
		},
	}

	s := newEmployeeManagerServiceForTest(repo)

	err := s.SelfAssignManager("cs1")
	if err == nil {
		t.Fatal("SelfAssignManager should fail when already self-assigned")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeEmployeeManagerExists) {
		t.Errorf("expected error code %q, got %v", domain.ErrorCodeEmployeeManagerExists, err)
	}
	if repo.addEmployeeManagersCalls != 0 {
		t.Errorf("AddEmployeeManagers should not be called on duplicate, got %d calls", repo.addEmployeeManagersCalls)
	}
}
