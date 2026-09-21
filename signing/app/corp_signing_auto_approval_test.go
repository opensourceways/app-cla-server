package app

import (
	"errors"
	"testing"
	"time"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
)

// fakeCorpSigningRepo embeds the repository.CorpSigning interface so only
// the methods used by the tests under consideration need to be overridden.
type fakeCorpSigningRepo struct {
	repository.CorpSigning

	findFn             func(index string) (domain.CorpSigning, error)
	updateAutoApproveFn func(cs *domain.CorpSigning) error
	addEmployeeFn      func(cs *domain.CorpSigning, es *domain.EmployeeSigning) error
	findByEmailFn     func(linkId string, email dp.EmailAddr) (repository.EmployeeSigningSummary, error)

	updateAutoApproveCalls int
	addEmployeeCalls      int
	lastAutoApprove       bool
}

func (f *fakeCorpSigningRepo) Find(index string) (domain.CorpSigning, error) {
	if f.findFn != nil {
		return f.findFn(index)
	}
	return domain.CorpSigning{}, nil
}

func (f *fakeCorpSigningRepo) UpdateAutoApprove(cs *domain.CorpSigning) error {
	f.updateAutoApproveCalls++
	f.lastAutoApprove = cs.AutoApproveEmployees
	if f.updateAutoApproveFn != nil {
		return f.updateAutoApproveFn(cs)
	}
	return nil
}

func (f *fakeCorpSigningRepo) AddEmployee(cs *domain.CorpSigning, es *domain.EmployeeSigning) error {
	f.addEmployeeCalls++
	if f.addEmployeeFn != nil {
		return f.addEmployeeFn(cs, es)
	}
	return nil
}

func (f *fakeCorpSigningRepo) FindEmployeesByEmail(linkId string, email dp.EmailAddr) (repository.EmployeeSigningSummary, error) {
	if f.findByEmailFn != nil {
		return f.findByEmailFn(linkId, email)
	}
	return repository.EmployeeSigningSummary{}, commonRepo.NewErrorResourceNotFound(errors.New("not found"))
}

// fakeVCService is a no-op verification code service for testing.
type fakeVCService struct {
	verifyErr error
}

func (f *fakeVCService) New(purpose dp.Purpose) (string, error) {
	return "code", nil
}

func (f *fakeVCService) NewIfItCan(purpose dp.Purpose, interval time.Duration) (string, error) {
	return "code", nil
}

func (f *fakeVCService) Verify(key *domain.VerificationCodeKey) error {
	return f.verifyErr
}

// ---- GetAutoApproval ----

func TestCorpSigningServiceGetAutoApproval(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"default off", false},
		{"enabled", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCorpSigningRepo{
				findFn: func(index string) (domain.CorpSigning, error) {
					return domain.CorpSigning{
						Id:                   index,
						AutoApproveEmployees: tt.enabled,
					}, nil
				},
			}

			s := newCorpSigningServiceForTest(repo)

			got, err := s.GetAutoApproval("cs1")
			if err != nil {
				t.Fatalf("GetAutoApproval: unexpected error %v", err)
			}
			if got != tt.enabled {
				t.Errorf("GetAutoApproval: got %v, want %v", got, tt.enabled)
			}
		})
	}
}

func TestCorpSigningServiceGetAutoApprovalFindError(t *testing.T) {
	findErr := errors.New("db error")
	repo := &fakeCorpSigningRepo{
		findFn: func(index string) (domain.CorpSigning, error) {
			return domain.CorpSigning{}, findErr
		},
	}

	s := newCorpSigningServiceForTest(repo)

	if _, err := s.GetAutoApproval("cs1"); err == nil {
		t.Error("GetAutoApproval should propagate Find error")
	}
}

// ---- UpdateAutoApproval ----

func TestCorpSigningServiceUpdateAutoApproval(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enable", true},
		{"disable", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCorpSigningRepo{
				findFn: func(index string) (domain.CorpSigning, error) {
					return domain.CorpSigning{Id: index, AutoApproveEmployees: !tt.enabled}, nil
				},
			}

			s := newCorpSigningServiceForTest(repo)

			if err := s.UpdateAutoApproval("cs1", tt.enabled); err != nil {
				t.Fatalf("UpdateAutoApproval: unexpected error %v", err)
			}
			if repo.updateAutoApproveCalls != 1 {
				t.Errorf("UpdateAutoApprove calls: got %d, want 1", repo.updateAutoApproveCalls)
			}
			if repo.lastAutoApprove != tt.enabled {
				t.Errorf("UpdateAutoApprove value: got %v, want %v", repo.lastAutoApprove, tt.enabled)
			}
		})
	}
}

func TestCorpSigningServiceUpdateAutoApprovalFindError(t *testing.T) {
	findErr := errors.New("db error")
	repo := &fakeCorpSigningRepo{
		findFn: func(index string) (domain.CorpSigning, error) {
			return domain.CorpSigning{}, findErr
		},
	}

	s := newCorpSigningServiceForTest(repo)

	if err := s.UpdateAutoApproval("cs1", true); err == nil {
		t.Error("UpdateAutoApproval should propagate Find error")
	}
	if repo.updateAutoApproveCalls != 0 {
		t.Errorf("UpdateAutoApprove should not be called when Find fails, got %d calls", repo.updateAutoApproveCalls)
	}
}

// ---- EmployeeSigningService.Sign auto-approval branch ----

func TestEmployeeSigningServiceSignAutoApprovalEnabled(t *testing.T) {
	email := dp.CreateEmailAddr("user@test.com")
	mgr := domain.Manager{
		Id:            "m1",
		Representative: domain.Representative{
			Name:      dp.CreateName("Manager"),
			EmailAddr: dp.CreateEmailAddr("mgr@test.com"),
		},
	}

	repo := &fakeCorpSigningRepo{
		findFn: func(index string) (domain.CorpSigning, error) {
			return domain.CorpSigning{
				Id:                   index,
				Link:                 domain.LinkInfo{Id: "link1"},
				Corp:                 domain.Corporation{AllEmailDomains: []string{"test.com"}},
				Managers:             []domain.Manager{mgr},
				AutoApproveEmployees: true,
			}, nil
		},
	}

	s := newEmployeeSigningServiceForTest(repo, &fakeVCService{})

	cmd := CmdToSignEmployeeCLA{
		CorpSigningId:    "cs1",
		Rep:              domain.Representative{EmailAddr: email},
		VerificationCode: "1234",
	}

	dtos, err := s.Sign(&cmd)
	if err != nil {
		t.Fatalf("Sign: unexpected error %v", err)
	}
	if len(dtos) != 1 {
		t.Fatalf("Sign returned %d manager DTOs, want 1", len(dtos))
	}
	if repo.addEmployeeCalls != 1 {
		t.Errorf("AddEmployee calls: got %d, want 1", repo.addEmployeeCalls)
	}
}

func TestEmployeeSigningServiceSignAutoApprovalDisabled(t *testing.T) {
	email := dp.CreateEmailAddr("user@test.com")
	mgr := domain.Manager{
		Id:            "m1",
		Representative: domain.Representative{
			Name:      dp.CreateName("Manager"),
			EmailAddr: dp.CreateEmailAddr("mgr@test.com"),
		},
	}

	repo := &fakeCorpSigningRepo{
		findFn: func(index string) (domain.CorpSigning, error) {
			return domain.CorpSigning{
				Id:                   index,
				Link:                 domain.LinkInfo{Id: "link1"},
				Corp:                 domain.Corporation{AllEmailDomains: []string{"test.com"}},
				Managers:             []domain.Manager{mgr},
				AutoApproveEmployees: false,
			}, nil
		},
	}

	s := newEmployeeSigningServiceForTest(repo, &fakeVCService{})

	cmd := CmdToSignEmployeeCLA{
		CorpSigningId:    "cs1",
		Rep:              domain.Representative{EmailAddr: email},
		VerificationCode: "1234",
	}

	if _, err := s.Sign(&cmd); err != nil {
		t.Fatalf("Sign: unexpected error %v", err)
	}
	if repo.addEmployeeCalls != 1 {
		t.Errorf("AddEmployee calls: got %d, want 1", repo.addEmployeeCalls)
	}
}

// helpers

func newCorpSigningServiceForTest(repo repository.CorpSigning) *corpSigningService {
	return &corpSigningService{repo: repo}
}

func newEmployeeSigningServiceForTest(repo repository.CorpSigning, vc vcservice.VCService) *employeeSigningService {
	return &employeeSigningService{
		repo:     repo,
		vc:       verificationCodeService{vc: vc},
		interval: time.Minute,
	}
}
