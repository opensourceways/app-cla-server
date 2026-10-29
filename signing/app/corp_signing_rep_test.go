package app

import (
	"errors"
	"testing"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain/userservice"
)

// fakeLinkRepo embeds repository.Link so only Find needs to be overridden.
type fakeLinkRepo struct {
	repository.Link

	findFn func(linkId string) (domain.Link, error)
}

func (f *fakeLinkRepo) Find(linkId string) (domain.Link, error) {
	if f.findFn != nil {
		return f.findFn(linkId)
	}
	return domain.Link{Id: linkId, Submitter: "cm-user"}, nil
}

// fakeUserService embeds userservice.UserService so only UpdateEmail
// needs to be overridden.
type fakeUserService struct {
	userservice.UserService

	updateEmailFn func(linkId, csId, adminId string, oldEmail, newEmail dp.EmailAddr) error

	updateEmailCalledWith struct {
		linkId   string
		csId     string
		adminId  string
		oldEmail string
		newEmail string
	}
	updateEmailCalled bool
}

func (f *fakeUserService) UpdateEmail(linkId, csId, adminId string, oldEmail, newEmail dp.EmailAddr) error {
	f.updateEmailCalled = true
	f.updateEmailCalledWith.linkId = linkId
	f.updateEmailCalledWith.csId = csId
	f.updateEmailCalledWith.adminId = adminId
	f.updateEmailCalledWith.oldEmail = oldEmail.EmailAddr()
	f.updateEmailCalledWith.newEmail = newEmail.EmailAddr()
	if f.updateEmailFn != nil {
		return f.updateEmailFn(linkId, csId, adminId, oldEmail, newEmail)
	}
	return nil
}

// fakeCorpSigningRepoForRep extends fakeCorpSigningRepo with Update support.
type fakeCorpSigningRepoForRep struct {
	fakeCorpSigningRepo

	updateFn func(cs *domain.CorpSigning) error

	updateCalledWith struct {
		adminName  string
		adminEmail string
	}
	updateCalled bool
}

func (f *fakeCorpSigningRepoForRep) Update(cs *domain.CorpSigning) error {
	f.updateCalled = true
	if cs.Admin.Id != "" {
		f.updateCalledWith.adminName = cs.Admin.Name.Name()
		f.updateCalledWith.adminEmail = cs.Admin.EmailAddr.EmailAddr()
	}
	if f.updateFn != nil {
		return f.updateFn(cs)
	}
	return nil
}

func newCorpSigningServiceForRepTest(
	repo *fakeCorpSigningRepoForRep,
	linkRepo repository.Link,
	userSvc userservice.UserService,
) *corpSigningService {
	return &corpSigningService{
		repo:        repo,
		linkRepo:    linkRepo,
		userService: userSvc,
	}
}

// ---- UpdateRepresentative ----

func TestUpdateRepresentativeSuccess(t *testing.T) {
	oldEmail := dp.CreateEmailAddr("rep@old.com")
	cs := domain.CorpSigning{
		Id:   "cs1",
		Link: domain.LinkInfo{Id: "link1"},
		Rep: domain.Representative{
			Name:      dp.CreateName("Old Name"),
			EmailAddr: oldEmail,
		},
		Admin: domain.Manager{
			Id: "admin1",
			Representative: domain.Representative{
				Name:      dp.CreateName("Old Name"),
				EmailAddr: oldEmail,
			},
		},
	}

	repo := &fakeCorpSigningRepoForRep{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return cs, nil
			},
		},
	}
	linkRepo := &fakeLinkRepo{}
	userSvc := &fakeUserService{}

	s := newCorpSigningServiceForRepTest(repo, linkRepo, userSvc)

	err := s.UpdateRepresentative("cm-user", "link1", "cs1", "New Name", "rep@new.com")
	if err != nil {
		t.Fatalf("UpdateRepresentative: unexpected error %v", err)
	}

	// Verify Update was called with synced admin info
	if !repo.updateCalled {
		t.Error("Update should be called")
	}
	if repo.updateCalledWith.adminName != "New Name" {
		t.Errorf("admin name: got %s, want New Name", repo.updateCalledWith.adminName)
	}
	if repo.updateCalledWith.adminEmail != "rep@new.com" {
		t.Errorf("admin email: got %s, want rep@new.com", repo.updateCalledWith.adminEmail)
	}

	// Verify UpdateEmail was called with csId and adminId
	if !userSvc.updateEmailCalled {
		t.Error("UpdateEmail should be called when admin exists")
	}
	if userSvc.updateEmailCalledWith.csId != "cs1" {
		t.Errorf("csId: got %s, want cs1", userSvc.updateEmailCalledWith.csId)
	}
	if userSvc.updateEmailCalledWith.adminId != "admin1" {
		t.Errorf("adminId: got %s, want admin1", userSvc.updateEmailCalledWith.adminId)
	}
	if userSvc.updateEmailCalledWith.oldEmail != "rep@old.com" {
		t.Errorf("oldEmail: got %s, want rep@old.com", userSvc.updateEmailCalledWith.oldEmail)
	}
	if userSvc.updateEmailCalledWith.newEmail != "rep@new.com" {
		t.Errorf("newEmail: got %s, want rep@new.com", userSvc.updateEmailCalledWith.newEmail)
	}
}

func TestUpdateRepresentativeNoAdminSkipsUpdateEmail(t *testing.T) {
	cs := domain.CorpSigning{
		Id:   "cs1",
		Link: domain.LinkInfo{Id: "link1"},
		Rep: domain.Representative{
			Name:      dp.CreateName("Old Name"),
			EmailAddr: dp.CreateEmailAddr("rep@old.com"),
		},
	}

	repo := &fakeCorpSigningRepoForRep{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return cs, nil
			},
		},
	}
	linkRepo := &fakeLinkRepo{}
	userSvc := &fakeUserService{}

	s := newCorpSigningServiceForRepTest(repo, linkRepo, userSvc)

	err := s.UpdateRepresentative("cm-user", "link1", "cs1", "New Name", "rep@new.com")
	if err != nil {
		t.Fatalf("UpdateRepresentative: unexpected error %v", err)
	}

	if userSvc.updateEmailCalled {
		t.Error("UpdateEmail should not be called when admin is empty")
	}
}

func TestUpdateRepresentativeLinkMismatch(t *testing.T) {
	cs := domain.CorpSigning{
		Id:   "cs1",
		Link: domain.LinkInfo{Id: "other-link"},
	}

	repo := &fakeCorpSigningRepoForRep{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return cs, nil
			},
		},
	}

	s := newCorpSigningServiceForRepTest(repo, &fakeLinkRepo{}, &fakeUserService{})

	err := s.UpdateRepresentative("cm-user", "link1", "cs1", "New Name", "rep@new.com")
	if err == nil {
		t.Fatal("UpdateRepresentative should fail when link_id does not match")
	}
}

func TestUpdateRepresentativeFindError(t *testing.T) {
	repo := &fakeCorpSigningRepoForRep{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return domain.CorpSigning{}, errors.New("db error")
			},
		},
	}

	s := newCorpSigningServiceForRepTest(repo, &fakeLinkRepo{}, &fakeUserService{})

	if err := s.UpdateRepresentative("cm-user", "link1", "cs1", "New Name", "rep@new.com"); err == nil {
		t.Error("UpdateRepresentative should propagate Find error")
	}
}

func TestUpdateRepresentativeNotCommunityManager(t *testing.T) {
	repo := &fakeCorpSigningRepoForRep{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return domain.CorpSigning{Id: "cs1", Link: domain.LinkInfo{Id: "link1"}}, nil
			},
		},
	}
	linkRepo := &fakeLinkRepo{
		findFn: func(linkId string) (domain.Link, error) {
			return domain.Link{Id: linkId, Submitter: "other-user"}, nil
		},
	}

	s := newCorpSigningServiceForRepTest(repo, linkRepo, &fakeUserService{})

	err := s.UpdateRepresentative("cm-user", "link1", "cs1", "New Name", "rep@new.com")
	if err == nil {
		t.Fatal("UpdateRepresentative should fail when user is not community manager")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeNoPermission) {
		t.Errorf("expected error code %q, got %v", domain.ErrorCodeNoPermission, err)
	}
}

func TestUpdateRepresentativeUpdateError(t *testing.T) {
	cs := domain.CorpSigning{
		Id:   "cs1",
		Link: domain.LinkInfo{Id: "link1"},
		Rep: domain.Representative{
			Name:      dp.CreateName("Old Name"),
			EmailAddr: dp.CreateEmailAddr("rep@old.com"),
		},
	}

	repo := &fakeCorpSigningRepoForRep{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return cs, nil
			},
		},
		updateFn: func(cs *domain.CorpSigning) error {
			return errors.New("update failed")
		},
	}

	s := newCorpSigningServiceForRepTest(repo, &fakeLinkRepo{}, &fakeUserService{})

	if err := s.UpdateRepresentative("cm-user", "link1", "cs1", "New Name", "rep@new.com"); err == nil {
		t.Error("UpdateRepresentative should propagate Update error")
	}
}

func TestUpdateRepresentativeUpdateEmailError(t *testing.T) {
	cs := domain.CorpSigning{
		Id:   "cs1",
		Link: domain.LinkInfo{Id: "link1"},
		Rep: domain.Representative{
			Name:      dp.CreateName("Old Name"),
			EmailAddr: dp.CreateEmailAddr("rep@old.com"),
		},
		Admin: domain.Manager{
			Id: "admin1",
			Representative: domain.Representative{
				Name:      dp.CreateName("Old Name"),
				EmailAddr: dp.CreateEmailAddr("rep@old.com"),
			},
		},
	}

	repo := &fakeCorpSigningRepoForRep{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return cs, nil
			},
		},
	}
	userSvc := &fakeUserService{
		updateEmailFn: func(linkId, csId, adminId string, oldEmail, newEmail dp.EmailAddr) error {
			return commonRepo.NewErrorResourceNotFound(errors.New("user not found"))
		},
	}

	s := newCorpSigningServiceForRepTest(repo, &fakeLinkRepo{}, userSvc)

	if err := s.UpdateRepresentative("cm-user", "link1", "cs1", "New Name", "rep@new.com"); err == nil {
		t.Error("UpdateRepresentative should propagate UpdateEmail error")
	}
}
