package app

import (
	"errors"
	"testing"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

// --- Fake implementations ---

type fakePDFStorage struct {
	uploadFn   func(corpSigningId string, data []byte) (string, error)
	downloadFn func(key string) ([]byte, error)

	uploadCalls   int
	downloadCalls int
	lastUploadID  string
	lastUploadData []byte
	lastDownloadKey string
}

func (f *fakePDFStorage) Upload(corpSigningId string, data []byte) (string, error) {
	f.uploadCalls++
	f.lastUploadID = corpSigningId
	f.lastUploadData = data
	if f.uploadFn != nil {
		return f.uploadFn(corpSigningId, data)
	}
	return corpSigningId + ".pdf", nil
}

func (f *fakePDFStorage) Download(key string) ([]byte, error) {
	f.downloadCalls++
	f.lastDownloadKey = key
	if f.downloadFn != nil {
		return f.downloadFn(key)
	}
	return []byte("fake-pdf-content"), nil
}

type fakeLinkRepoForPDF struct {
	findFn func(string) (domain.Link, error)
}

func (f *fakeLinkRepoForPDF) Find(linkId string) (domain.Link, error) {
	if f.findFn != nil {
		return f.findFn(linkId)
	}
	return domain.Link{Id: linkId, Submitter: "user1"}, nil
}

// Embed the interface to satisfy all methods with nil defaults
func (f *fakeLinkRepoForPDF) NewLinkId() string { return "link1" }
func (f *fakeLinkRepoForPDF) Add(*domain.Link) error { return nil }
func (f *fakeLinkRepoForPDF) Remove(*domain.Link) error { return nil }
func (f *fakeLinkRepoForPDF) FindAll(string) ([]repository.LinkSummary, error) { return nil, nil }
func (f *fakeLinkRepoForPDF) ListAll() ([]repository.LinkCLA, error) { return nil, nil }
func (f *fakeLinkRepoForPDF) AddCLA(*domain.Link, *domain.CLA) error { return nil }
func (f *fakeLinkRepoForPDF) UpdateCLA(*domain.Link, *domain.CLA) error { return nil }
func (f *fakeLinkRepoForPDF) RemoveCLA(*domain.Link, *domain.CLA) error { return nil }
func (f *fakeLinkRepoForPDF) UpdateGracePeriodDays(string, int) error { return nil }

// fakeCorpSigningRepoForPDF extends the existing fakeCorpSigningRepo
// with PDF-specific methods
type fakeCorpSigningRepoForPDF struct {
	fakeCorpSigningRepo

	saveCorpPDFURLFn            func(cs *domain.CorpSigning, url string) error
	findCorpPDFURLFn            func(csId string) (string, error)
	findCorpPDFFn               func(csId string) ([]byte, error)
	setCorpPDFURLAndUnsetPDFFn  func(csId, url string) error
	findCorpSigningIDsForMigFn  func(page, pageSize int) ([]string, int64, error)

	saveCorpPDFURLCalls           int
	lastSavedURL                  string
	findCorpPDFURLCalls           int
	findCorpPDFFnCalls            int
	setCorpPDFURLAndUnsetPDFCalls int
	findMigCalls                  int
}

func (f *fakeCorpSigningRepoForPDF) SaveCorpPDFURL(cs *domain.CorpSigning, url string) error {
	f.saveCorpPDFURLCalls++
	f.lastSavedURL = url
	if f.saveCorpPDFURLFn != nil {
		return f.saveCorpPDFURLFn(cs, url)
	}
	return nil
}

func (f *fakeCorpSigningRepoForPDF) FindCorpPDFURL(csId string) (string, error) {
	f.findCorpPDFURLCalls++
	if f.findCorpPDFURLFn != nil {
		return f.findCorpPDFURLFn(csId)
	}
	return "", nil
}

func (f *fakeCorpSigningRepoForPDF) FindCorpPDF(csId string) ([]byte, error) {
	f.findCorpPDFFnCalls++
	if f.findCorpPDFFn != nil {
		return f.findCorpPDFFn(csId)
	}
	return nil, nil
}

func (f *fakeCorpSigningRepoForPDF) SetCorpPDFURLAndUnsetPDF(csId, url string) error {
	f.setCorpPDFURLAndUnsetPDFCalls++
	if f.setCorpPDFURLAndUnsetPDFFn != nil {
		return f.setCorpPDFURLAndUnsetPDFFn(csId, url)
	}
	return nil
}

func (f *fakeCorpSigningRepoForPDF) FindCorpSigningIDsForPDFMigration(page, pageSize int) ([]string, int64, error) {
	f.findMigCalls++
	if f.findCorpSigningIDsForMigFn != nil {
		return f.findCorpSigningIDsForMigFn(page, pageSize)
	}
	return nil, 0, nil
}

// --- Helpers ---

func newCorpPDFServiceForTest(
	repo *fakeCorpSigningRepoForPDF,
	linkRepo *fakeLinkRepoForPDF,
	obs *fakePDFStorage,
) *corpPDFService {
	return &corpPDFService{
		repo:     repo,
		linkRepo: linkRepo,
		obs:      obs,
	}
}

func defaultUploadCmd() *CmdToUploadCorpSigningPDF {
	return &CmdToUploadCorpSigningPDF{
		UserId: "user1",
		CSId:   "cs1",
		PDF:    []byte("pdf-bytes"),
	}
}

func defaultFindCSRepo() *fakeCorpSigningRepoForPDF {
	return &fakeCorpSigningRepoForPDF{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return domain.CorpSigning{
					Id:   index,
					Link: domain.LinkInfo{Id: "link1"},
				}, nil
			},
		},
	}
}

func defaultLinkRepo() *fakeLinkRepoForPDF {
	return &fakeLinkRepoForPDF{
		findFn: func(linkId string) (domain.Link, error) {
			return domain.Link{Id: linkId, Submitter: "user1"}, nil
		},
	}
}

// --- Upload tests ---

func TestCorpPDFServiceUpload(t *testing.T) {
	repo := defaultFindCSRepo()
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	err := s.Upload(defaultUploadCmd())
	if err != nil {
		t.Fatalf("Upload: unexpected error %v", err)
	}
	if obs.uploadCalls != 1 {
		t.Errorf("Upload: obs upload calls got %d, want 1", obs.uploadCalls)
	}
	if repo.saveCorpPDFURLCalls != 1 {
		t.Errorf("Upload: repo SaveCorpPDFURL calls got %d, want 1", repo.saveCorpPDFURLCalls)
	}
	if repo.lastSavedURL != "cs1.pdf" {
		t.Errorf("Upload: saved URL got %q, want %q", repo.lastSavedURL, "cs1.pdf")
	}
}

func TestCorpPDFServiceUploadOBSFails(t *testing.T) {
	repo := defaultFindCSRepo()
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{
		uploadFn: func(string, []byte) (string, error) {
			return "", errors.New("obs error")
		},
	}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	err := s.Upload(defaultUploadCmd())
	if err == nil {
		t.Fatal("Upload: expected error when OBS upload fails")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeCorpPDFUploadFailed) {
		t.Errorf("Upload: expected error code %q, got %v", domain.ErrorCodeCorpPDFUploadFailed, err)
	}
	if repo.saveCorpPDFURLCalls != 0 {
		t.Errorf("Upload: should not save URL when OBS upload fails, got %d calls", repo.saveCorpPDFURLCalls)
	}
}

func TestCorpPDFServiceUploadFindFails(t *testing.T) {
	repo := &fakeCorpSigningRepoForPDF{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(string) (domain.CorpSigning, error) {
				return domain.CorpSigning{}, errors.New("db error")
			},
		},
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	err := s.Upload(defaultUploadCmd())
	if err == nil {
		t.Fatal("Upload: expected error when Find fails")
	}
	if obs.uploadCalls != 0 {
		t.Errorf("Upload: should not call OBS when Find fails, got %d calls", obs.uploadCalls)
	}
}

func TestCorpPDFServiceUploadNotCommunityManager(t *testing.T) {
	repo := defaultFindCSRepo()
	linkRepo := &fakeLinkRepoForPDF{
		findFn: func(linkId string) (domain.Link, error) {
			return domain.Link{Id: linkId, Submitter: "other-user"}, nil
		},
	}
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	err := s.Upload(defaultUploadCmd())
	if err == nil {
		t.Fatal("Upload: expected error when user is not community manager")
	}
	if obs.uploadCalls != 0 {
		t.Errorf("Upload: should not call OBS when not community manager, got %d calls", obs.uploadCalls)
	}
}

// --- Download tests ---

func TestCorpPDFServiceDownloadFromOBS(t *testing.T) {
	repo := defaultFindCSRepo()
	repo.findCorpPDFURLFn = func(string) (string, error) {
		return "cs1.pdf", nil
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{
		downloadFn: func(key string) ([]byte, error) {
			return []byte("pdf-from-obs"), nil
		},
	}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	pdf, err := s.Download("user1", "cs1", dp.CreateEmailAddr("admin@test.com"))
	if err != nil {
		t.Fatalf("Download: unexpected error %v", err)
	}
	if string(pdf) != "pdf-from-obs" {
		t.Errorf("Download: got %q, want %q", string(pdf), "pdf-from-obs")
	}
	if repo.findCorpPDFURLCalls != 1 {
		t.Errorf("Download: FindCorpPDFURL calls got %d, want 1", repo.findCorpPDFURLCalls)
	}
	if obs.downloadCalls != 1 {
		t.Errorf("Download: OBS download calls got %d, want 1", obs.downloadCalls)
	}
	if repo.findCorpPDFFnCalls != 0 {
		t.Errorf("Download: should not call FindCorpPDF when URL exists, got %d calls", repo.findCorpPDFFnCalls)
	}
}

func TestCorpPDFServiceDownloadLegacyFallback(t *testing.T) {
	repo := defaultFindCSRepo()
	repo.findCorpPDFURLFn = func(string) (string, error) {
		return "", nil
	}
	repo.findCorpPDFFn = func(string) ([]byte, error) {
		return []byte("legacy-pdf"), nil
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	pdf, err := s.Download("user1", "cs1", dp.CreateEmailAddr("admin@test.com"))
	if err != nil {
		t.Fatalf("Download: unexpected error %v", err)
	}
	if string(pdf) != "legacy-pdf" {
		t.Errorf("Download: got %q, want %q", string(pdf), "legacy-pdf")
	}
	if obs.downloadCalls != 0 {
		t.Errorf("Download: should not call OBS for legacy data, got %d calls", obs.downloadCalls)
	}
	if repo.findCorpPDFFnCalls != 1 {
		t.Errorf("Download: FindCorpPDF calls got %d, want 1", repo.findCorpPDFFnCalls)
	}
}

func TestCorpPDFServiceDownloadNotFound(t *testing.T) {
	repo := defaultFindCSRepo()
	repo.findCorpPDFURLFn = func(string) (string, error) {
		return "", commonRepo.NewErrorResourceNotFound(errors.New("not found"))
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	_, err := s.Download("user1", "cs1", dp.CreateEmailAddr("admin@test.com"))
	if err == nil {
		t.Fatal("Download: expected error when PDF not found")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeCorpPDFNotFound) {
		t.Errorf("Download: expected error code %q, got %v", domain.ErrorCodeCorpPDFNotFound, err)
	}
}

func TestCorpPDFServiceDownloadLegacyEmptyPDF(t *testing.T) {
	repo := defaultFindCSRepo()
	repo.findCorpPDFURLFn = func(string) (string, error) {
		return "", nil
	}
	repo.findCorpPDFFn = func(string) ([]byte, error) {
		return []byte{}, nil
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	_, err := s.Download("user1", "cs1", dp.CreateEmailAddr("admin@test.com"))
	if err == nil {
		t.Fatal("Download: expected error when legacy PDF is empty")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeCorpPDFNotFound) {
		t.Errorf("Download: expected error code %q, got %v", domain.ErrorCodeCorpPDFNotFound, err)
	}
}

func TestCorpPDFServiceDownloadOBSFails(t *testing.T) {
	repo := defaultFindCSRepo()
	repo.findCorpPDFURLFn = func(string) (string, error) {
		return "cs1.pdf", nil
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{
		downloadFn: func(string) ([]byte, error) {
			return nil, errors.New("obs download error")
		},
	}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	_, err := s.Download("user1", "cs1", dp.CreateEmailAddr("admin@test.com"))
	if err == nil {
		t.Fatal("Download: expected error when OBS download fails")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeCorpPDFNotFound) {
		t.Errorf("Download: expected error code %q, got %v", domain.ErrorCodeCorpPDFNotFound, err)
	}
}

func TestCorpPDFServiceDownloadAsAdmin(t *testing.T) {
	repo := &fakeCorpSigningRepoForPDF{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(index string) (domain.CorpSigning, error) {
				return domain.CorpSigning{
					Id:   index,
					Link: domain.LinkInfo{Id: "link1"},
					Admin: domain.Manager{
						Representative: domain.Representative{
							EmailAddr: dp.CreateEmailAddr("admin@test.com"),
						},
					},
				}, nil
			},
		},
		findCorpPDFURLFn: func(string) (string, error) {
			return "cs1.pdf", nil
		},
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{
		downloadFn: func(string) ([]byte, error) {
			return []byte("admin-pdf"), nil
		},
	}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	pdf, err := s.Download("user1", "cs1", dp.CreateEmailAddr("admin@test.com"))
	if err != nil {
		t.Fatalf("Download as admin: unexpected error %v", err)
	}
	if string(pdf) != "admin-pdf" {
		t.Errorf("Download as admin: got %q, want %q", string(pdf), "admin-pdf")
	}
}

// --- Migration tests ---

func TestMigrateCorpPDFsToOBS(t *testing.T) {
	repo := defaultFindCSRepo()
	called := false
	repo.findCorpSigningIDsForMigFn = func(page, pageSize int) ([]string, int64, error) {
		if called {
			return nil, 0, nil
		}
		called = true
		return []string{"cs1", "cs2"}, 2, nil
	}
	repo.findCorpPDFFn = func(csId string) ([]byte, error) {
		return []byte("pdf-" + csId), nil
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	err := s.MigrateCorpPDFsToOBS(50)
	if err != nil {
		t.Fatalf("MigrateCorpPDFsToOBS: unexpected error %v", err)
	}
	if obs.uploadCalls != 2 {
		t.Errorf("MigrateCorpPDFsToOBS: OBS upload calls got %d, want 2", obs.uploadCalls)
	}
	if repo.setCorpPDFURLAndUnsetPDFCalls != 2 {
		t.Errorf("MigrateCorpPDFsToOBS: SetCorpPDFURLAndUnsetPDF calls got %d, want 2", repo.setCorpPDFURLAndUnsetPDFCalls)
	}
}

func TestMigrateCorpPDFsToOBSNoData(t *testing.T) {
	repo := defaultFindCSRepo()
	repo.findCorpSigningIDsForMigFn = func(page, pageSize int) ([]string, int64, error) {
		return nil, 0, nil
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	err := s.MigrateCorpPDFsToOBS(50)
	if err != nil {
		t.Fatalf("MigrateCorpPDFsToOBS: unexpected error %v", err)
	}
	if obs.uploadCalls != 0 {
		t.Errorf("MigrateCorpPDFsToOBS: should not call OBS when no data, got %d calls", obs.uploadCalls)
	}
}

func TestMigrateCorpPDFsToOBSPartialFailure(t *testing.T) {
	repo := defaultFindCSRepo()
	called := false
	repo.findCorpSigningIDsForMigFn = func(page, pageSize int) ([]string, int64, error) {
		if called {
			return nil, 0, nil
		}
		called = true
		return []string{"cs1", "cs2"}, 2, nil
	}
	repo.findCorpPDFFn = func(csId string) ([]byte, error) {
		if csId == "cs1" {
			return []byte("pdf-1"), nil
		}
		return nil, errors.New("read error")
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	err := s.MigrateCorpPDFsToOBS(50)
	if err != nil {
		t.Fatalf("MigrateCorpPDFsToOBS: should not fail on partial errors, got %v", err)
	}
	if obs.uploadCalls != 1 {
		t.Errorf("MigrateCorpPDFsToOBS: OBS upload calls got %d, want 1 (only cs1 succeeds)", obs.uploadCalls)
	}
}

func TestMigrateCorpPDFsToOBSEmptyPDF(t *testing.T) {
	repo := defaultFindCSRepo()
	called := false
	repo.findCorpSigningIDsForMigFn = func(page, pageSize int) ([]string, int64, error) {
		if called {
			return nil, 0, nil
		}
		called = true
		return []string{"cs1"}, 1, nil
	}
	repo.findCorpPDFFn = func(string) ([]byte, error) {
		return []byte{}, nil
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	err := s.MigrateCorpPDFsToOBS(50)
	if err != nil {
		t.Fatalf("MigrateCorpPDFsToOBS: unexpected error %v", err)
	}
	if obs.uploadCalls != 0 {
		t.Errorf("MigrateCorpPDFsToOBS: should not upload empty PDF, got %d calls", obs.uploadCalls)
	}
	if repo.setCorpPDFURLAndUnsetPDFCalls != 0 {
		t.Errorf("MigrateCorpPDFsToOBS: should not set URL for empty PDF, got %d calls", repo.setCorpPDFURLAndUnsetPDFCalls)
	}
}

func TestMigrateCorpPDFsToOBSFindIDsFails(t *testing.T) {
	repo := defaultFindCSRepo()
	repo.findCorpSigningIDsForMigFn = func(page, pageSize int) ([]string, int64, error) {
		return nil, 0, errors.New("db error")
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	err := s.MigrateCorpPDFsToOBS(50)
	if err == nil {
		t.Fatal("MigrateCorpPDFsToOBS: expected error when FindCorpSigningIDsForPDFMigration fails")
	}
}

func TestMigrateCorpPDFsToOBSFindPDFNotFound(t *testing.T) {
	repo := defaultFindCSRepo()
	called := false
	repo.findCorpSigningIDsForMigFn = func(page, pageSize int) ([]string, int64, error) {
		if called {
			return nil, 0, nil
		}
		called = true
		return []string{"cs1"}, 1, nil
	}
	repo.findCorpPDFFn = func(string) ([]byte, error) {
		return nil, commonRepo.NewErrorResourceNotFound(errors.New("not found"))
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	err := s.MigrateCorpPDFsToOBS(50)
	if err != nil {
		t.Fatalf("MigrateCorpPDFsToOBS: should skip not found, got %v", err)
	}
	if obs.uploadCalls != 0 {
		t.Errorf("MigrateCorpPDFsToOBS: should not upload when PDF not found, got %d calls", obs.uploadCalls)
	}
}

// --- Additional coverage tests ---

func TestCorpPDFServiceDownloadNotManager(t *testing.T) {
	repo := defaultFindCSRepo()
	repo.findCorpPDFURLFn = func(string) (string, error) {
		return "cs1.pdf", nil
	}
	linkRepo := &fakeLinkRepoForPDF{
		findFn: func(linkId string) (domain.Link, error) {
			return domain.Link{Id: linkId, Submitter: "other-user"}, nil
		},
	}
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	_, err := s.Download("user1", "cs1", dp.CreateEmailAddr("nobody@test.com"))
	if err == nil {
		t.Fatal("Download: expected error when user is not admin or manager")
	}
}

func TestCorpPDFServiceDownloadFindCorpPDFURLError(t *testing.T) {
	repo := defaultFindCSRepo()
	repo.findCorpPDFURLFn = func(string) (string, error) {
		return "", errors.New("db connection error")
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	_, err := s.Download("user1", "cs1", dp.CreateEmailAddr("admin@test.com"))
	if err == nil {
		t.Fatal("Download: expected error when FindCorpPDFURL returns non-not-found error")
	}
}

func TestCorpPDFServiceDownloadLinkNotFound(t *testing.T) {
	repo := defaultFindCSRepo()
	repo.findCorpPDFURLFn = func(string) (string, error) {
		return "cs1.pdf", nil
	}
	linkRepo := &fakeLinkRepoForPDF{
		findFn: func(string) (domain.Link, error) {
			return domain.Link{}, commonRepo.NewErrorResourceNotFound(errors.New("not found"))
		},
	}
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	_, err := s.Download("user1", "cs1", dp.CreateEmailAddr("nobody@test.com"))
	if err == nil {
		t.Fatal("Download: expected error when link not found")
	}
	if !domain.IsErrorOf(err, domain.ErrorCodeLinkNotExists) {
		t.Errorf("Download: expected error code %q, got %v", domain.ErrorCodeLinkNotExists, err)
	}
}

func TestCorpPDFServiceDownloadLinkFindError(t *testing.T) {
	repo := defaultFindCSRepo()
	repo.findCorpPDFURLFn = func(string) (string, error) {
		return "cs1.pdf", nil
	}
	linkRepo := &fakeLinkRepoForPDF{
		findFn: func(string) (domain.Link, error) {
			return domain.Link{}, errors.New("link db error")
		},
	}
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	_, err := s.Download("user1", "cs1", dp.CreateEmailAddr("nobody@test.com"))
	if err == nil {
		t.Fatal("Download: expected error when link repo returns error")
	}
}

func TestCorpPDFServiceDownloadFindFails(t *testing.T) {
	repo := &fakeCorpSigningRepoForPDF{
		fakeCorpSigningRepo: fakeCorpSigningRepo{
			findFn: func(string) (domain.CorpSigning, error) {
				return domain.CorpSigning{}, errors.New("db error")
			},
		},
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	_, err := s.Download("user1", "cs1", dp.CreateEmailAddr("admin@test.com"))
	if err == nil {
		t.Fatal("Download: expected error when Find fails")
	}
}

func TestCorpPDFServiceDownloadLegacyFindError(t *testing.T) {
	repo := defaultFindCSRepo()
	repo.findCorpPDFURLFn = func(string) (string, error) {
		return "", nil
	}
	repo.findCorpPDFFn = func(string) ([]byte, error) {
		return nil, errors.New("legacy db error")
	}
	linkRepo := defaultLinkRepo()
	obs := &fakePDFStorage{}

	s := newCorpPDFServiceForTest(repo, linkRepo, obs)

	_, err := s.Download("user1", "cs1", dp.CreateEmailAddr("admin@test.com"))
	if err == nil {
		t.Fatal("Download: expected error when legacy FindCorpPDF returns error")
	}
}
