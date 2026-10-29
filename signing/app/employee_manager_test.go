package app

import (
	"errors"
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

// fakeEmployeeManagerRepo embeds repository.CorpSigning so only the methods
// used by EmployeeManagerService tests need to be overridden.
type fakeEmployeeManagerRepo struct {
	repository.CorpSigning

	findFn                func(index string) (domain.CorpSigning, error)
	addEmployeeManagersFn func(cs *domain.CorpSigning, ms []domain.Manager) error

	addEmployeeManagersCalls int
	lastAddedManagers         []domain.Manager
}

func (f *fakeEmployeeManagerRepo) Find(index string) (domain.CorpSigning, error) {
	if f.findFn != nil {
		return f.findFn(index)
	}
	return domain.CorpSigning{}, nil
}

func (f *fakeEmployeeManagerRepo) AddEmployeeManagers(cs *domain.CorpSigning, ms []domain.Manager) error {
	f.addEmployeeManagersCalls++
	f.lastAddedManagers = ms
	if f.addEmployeeManagersFn != nil {
		return f.addEmployeeManagersFn(cs, ms)
	}
	return nil
}

func newEmployeeManagerServiceForTest(repo repository.CorpSigning) *employeeManagerService {
	return &employeeManagerService{
		repo:        repo,
		userService: nil,
	}
}

func TestSelfAssignManagerSuccess(t *testing.T) {
	adminEmail := dp.CreateEmailAddr("admin@test.com")
	repo := &fakeEmployeeManagerRepo{
		findFn: func(index string) (domain.CorpSigning, error) {
			return domain.CorpSigning{
				Id:   index,
				Corp: domain.Corporation{AllEmailDomains: []string{"test.com"}},
				Admin: domain.Manager{
					Id:            "admin1",
					Representative: domain.Representative{
						Name:      dp.CreateName("Admin"),
						EmailAddr: adminEmail,
					},
				},
			}, nil
		},
	}

	s := newEmployeeManagerServiceForTest(repo)

	if err := s.SelfAssignManager("cs1"); err != nil {
		t.Fatalf("SelfAssignManager: unexpected error %v", err)
	}

	if repo.addEmployeeManagersCalls != 1 {
		t.Errorf("AddEmployeeManagers calls: got %d, want 1", repo.addEmployeeManagersCalls)
	}

	if len(repo.lastAddedManagers) != 1 {
		t.Fatalf("lastAddedManagers length: got %d, want 1", len(repo.lastAddedManagers))
	}

	m := repo.lastAddedManagers[0]
	if m.Id != "self" {
		t.Errorf("Manager Id: got %s, want self", m.Id)
	}
	if m.EmailAddr.EmailAddr() != "admin@test.com" {
		t.Errorf("Manager Email: got %s, want admin@test.com", m.EmailAddr.EmailAddr())
	}
}

func TestSelfAssignManagerAdminNotExists(t *testing.T) {
	repo := &fakeEmployeeManagerRepo{
		findFn: func(index string) (domain.CorpSigning, error) {
			return domain.CorpSigning{
				Id:    index,
				Corp:  domain.Corporation{AllEmailDomains: []string{"test.com"}},
				Admin: domain.Manager{},
			}, nil
		},
	}

	s := newEmployeeManagerServiceForTest(repo)

	err := s.SelfAssignManager("cs1")
	if err == nil {
		t.Fatal("SelfAssignManager should fail when admin is empty")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeCorpAdminNotExists) {
		t.Errorf("expected error code %s, got %v", domain.ErrorCodeCorpAdminNotExists, err)
	}
	if repo.addEmployeeManagersCalls != 0 {
		t.Errorf("AddEmployeeManagers should not be called, got %d calls", repo.addEmployeeManagersCalls)
	}
}

func TestSelfAssignManagerFindError(t *testing.T) {
	findErr := errors.New("db error")
	repo := &fakeEmployeeManagerRepo{
		findFn: func(index string) (domain.CorpSigning, error) {
			return domain.CorpSigning{}, findErr
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

func TestSelfAssignManagerDuplicate(t *testing.T) {
	adminEmail := dp.CreateEmailAddr("admin@test.com")
	repo := &fakeEmployeeManagerRepo{
		findFn: func(index string) (domain.CorpSigning, error) {
			return domain.CorpSigning{
				Id:   index,
				Corp: domain.Corporation{AllEmailDomains: []string{"test.com"}},
				Admin: domain.Manager{
					Id:            "admin1",
					Representative: domain.Representative{
						Name:      dp.CreateName("Admin"),
						EmailAddr: adminEmail,
					},
				},
				Managers: []domain.Manager{
					{
						Id:            "self",
						Representative: domain.Representative{
							Name:      dp.CreateName("Admin"),
							EmailAddr: adminEmail,
						},
					},
				},
			}, nil
		},
	}

	s := newEmployeeManagerServiceForTest(repo)

	err := s.SelfAssignManager("cs1")
	if err == nil {
		t.Fatal("SelfAssignManager should fail when already self-assigned")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeEmployeeManagerExists) {
		t.Errorf("expected error code %s, got %v", domain.ErrorCodeEmployeeManagerExists, err)
	}
	if repo.addEmployeeManagersCalls != 0 {
		t.Errorf("AddEmployeeManagers should not be called on duplicate, got %d calls", repo.addEmployeeManagersCalls)
	}
}

func TestSelfAssignManagerTooMany(t *testing.T) {
	adminEmail := dp.CreateEmailAddr("admin@test.com")
	managers := make([]domain.Manager, 5)
	for i := range managers {
		managers[i] = domain.Manager{
			Id:            "m" + string(rune('a'+i)),
			Representative: domain.Representative{
				Name:      dp.CreateName("Mgr"),
				EmailAddr: dp.CreateEmailAddr("mgr"+string(rune('a'+i))+"@test.com"),
			},
		}
	}

	repo := &fakeEmployeeManagerRepo{
		findFn: func(index string) (domain.CorpSigning, error) {
			return domain.CorpSigning{
				Id:       index,
				Corp:     domain.Corporation{AllEmailDomains: []string{"test.com"}},
				Managers: managers,
				Admin: domain.Manager{
					Id:            "admin1",
					Representative: domain.Representative{
						Name:      dp.CreateName("Admin"),
						EmailAddr: adminEmail,
					},
				},
			}, nil
		},
	}

	s := newEmployeeManagerServiceForTest(repo)

	err := s.SelfAssignManager("cs1")
	if err == nil {
		t.Fatal("SelfAssignManager should fail when managers exceed limit")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeEmployeeManagerTooMany) {
		t.Errorf("expected error code %s, got %v", domain.ErrorCodeEmployeeManagerTooMany, err)
	}
	if repo.addEmployeeManagersCalls != 0 {
		t.Errorf("AddEmployeeManagers should not be called when too many, got %d calls", repo.addEmployeeManagersCalls)
	}
}

func TestSelfAssignManagerAddEmployeeManagersError(t *testing.T) {
	adminEmail := dp.CreateEmailAddr("admin@test.com")
	repo := &fakeEmployeeManagerRepo{
		findFn: func(index string) (domain.CorpSigning, error) {
			return domain.CorpSigning{
				Id:   index,
				Corp: domain.Corporation{AllEmailDomains: []string{"test.com"}},
				Admin: domain.Manager{
					Id:            "admin1",
					Representative: domain.Representative{
						Name:      dp.CreateName("Admin"),
						EmailAddr: adminEmail,
					},
				},
			}, nil
		},
		addEmployeeManagersFn: func(cs *domain.CorpSigning, ms []domain.Manager) error {
			return errors.New("db write error")
		},
	}

	s := newEmployeeManagerServiceForTest(repo)

	err := s.SelfAssignManager("cs1")
	if err == nil {
		t.Fatal("SelfAssignManager should propagate AddEmployeeManagers error")
	}
	if repo.addEmployeeManagersCalls != 1 {
		t.Errorf("AddEmployeeManagers should be called once, got %d calls", repo.addEmployeeManagersCalls)
	}
}
