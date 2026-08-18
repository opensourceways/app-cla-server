package app

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/claservice"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func TestMain(m *testing.M) {
	dp.Init(&dp.Config{
		MaxLengthOfName:     100,
		MaxLengthOfTitle:    200,
		MaxLengthOfEmail:    200,
		MaxLengthOfAccount:  100,
		MaxLengthOfCorpName: 200,
	})

	os.Exit(m.Run())
}

// ---- fakes ----

type fakeCLAConfirmToken struct {
	repository.CLAConfirmToken

	payload repository.CLAConfirmTokenPayload
	err     error
}

func (f *fakeCLAConfirmToken) Consume(token string) (repository.CLAConfirmTokenPayload, error) {
	if f.err != nil {
		return repository.CLAConfirmTokenPayload{}, f.err
	}

	return f.payload, nil
}

type fakeIndividualSigningRepo struct {
	repository.IndividualSigning

	sign       domain.IndividualSigning
	findErr    error
	saved      *domain.IndividualSigning
	saveErr    error
	saveCalled bool
}

func (f *fakeIndividualSigningRepo) Find(linkId string, email dp.EmailAddr) (domain.IndividualSigning, error) {
	if f.findErr != nil {
		return domain.IndividualSigning{}, f.findErr
	}

	return f.sign, nil
}

func (f *fakeIndividualSigningRepo) SaveNewCLA(is *domain.IndividualSigning) error {
	f.saveCalled = true
	if f.saveErr != nil {
		return f.saveErr
	}

	f.saved = is

	return nil
}

type fakeCLAService struct {
	claservice.CLAService
	claId string
}

func (f *fakeCLAService) GetClaId(linkId string, claType dp.CLAType, language dp.Language) string {
	return f.claId
}

type fakeLinkRepo struct {
	repository.Link

	link    domain.Link
	findErr error
}

func (f *fakeLinkRepo) Find(linkId string) (domain.Link, error) {
	if f.findErr != nil {
		return domain.Link{}, f.findErr
	}

	return f.link, nil
}

func newServiceForConfirm(
	token repository.CLAConfirmToken,
	repo repository.IndividualSigning,
	cla claservice.CLAService,
	linkRepo repository.Link,
) *individualSigningService {
	return &individualSigningService{
		cla:             cla,
		repo:            repo,
		linkRepo:        linkRepo,
		claConfirmToken: token,
	}
}

func signedIndividual(claId string) domain.IndividualSigning {
	return domain.IndividualSigning{
		Link: domain.LinkInfo{
			Id: "link1",
			CLAInfo: domain.CLAInfo{
				CLAId:    claId,
				Language: dp.CreateLanguage("en"),
			},
		},
		Rep: domain.Representative{
			EmailAddr: dp.CreateEmailAddr("alice@example.com"),
		},
	}
}

// ---- tests ----

func TestConfirmByTokenSuccess(t *testing.T) {
	token := &fakeCLAConfirmToken{
		payload: repository.CLAConfirmTokenPayload{
			LinkId:   "link1",
			Email:    "alice@example.com",
			NewCLAId: "8",
		},
	}
	repo := &fakeIndividualSigningRepo{sign: signedIndividual("7")}
	cla := &fakeCLAService{claId: "9"}
	linkRepo := &fakeLinkRepo{link: domain.Link{Id: "link1", Org: domain.OrgInfo{Alias: "openeuler"}}}

	dto, err := newServiceForConfirm(token, repo, cla, linkRepo).ConfirmByToken(&CmdToConfirmCLAByToken{Token: "t"})
	if err != nil {
		t.Fatalf("ConfirmByToken failed: %v", err)
	}

	if !repo.saveCalled {
		t.Fatal("SaveNewCLA was not called")
	}
	if repo.saved.Link.CLAId != "9" {
		t.Fatalf("saved cla id = %s, want the latest one (9)", repo.saved.Link.CLAId)
	}

	if dto.Result != "confirmed" {
		t.Fatalf("result = %s, want confirmed", dto.Result)
	}
	if dto.OrgAlias != "openeuler" {
		t.Fatalf("org alias = %s, want openeuler", dto.OrgAlias)
	}
	if dto.EmailMasked != "alic***@example.com" {
		t.Fatalf("masked email = %s", dto.EmailMasked)
	}
	if dto.LinkId != "link1" || dto.ClaId != "9" {
		t.Fatalf("audit context = %+v", dto)
	}

	// the agree log must be appended
	logs := repo.saved.Logs
	if len(logs) == 0 || logs[len(logs)-1].Action != "agree" || logs[len(logs)-1].ClaId != "9" {
		t.Fatalf("unexpected logs: %+v", logs)
	}
}

func TestConfirmByTokenInvalidTokenVariants(t *testing.T) {
	notFound := commonRepo.NewErrorResourceNotFound(errors.New("not exists"))
	formatInvalid := domain.NewDomainError(domain.ErrorCodeCLAConfirmTokenInvalid)

	cases := []struct {
		name  string
		token repository.CLAConfirmToken
	}{
		{"unknown or expired token", &fakeCLAConfirmToken{err: notFound}},
		{"malformed token", &fakeCLAConfirmToken{err: formatInvalid}},
		{"payload email invalid", &fakeCLAConfirmToken{
			payload: repository.CLAConfirmTokenPayload{LinkId: "link1", Email: "not-an-email"},
		}},
		{"signing record not exists", &fakeCLAConfirmToken{
			payload: repository.CLAConfirmTokenPayload{LinkId: "link1", Email: "bob@example.com"},
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := &fakeIndividualSigningRepo{findErr: commonRepo.NewErrorResourceNotFound(errors.New("no record"))}
			s := newServiceForConfirm(c.token, repo, &fakeCLAService{claId: "9"}, &fakeLinkRepo{})

			_, err := s.ConfirmByToken(&CmdToConfirmCLAByToken{Token: "t"})
			if !domain.IsErrorOf(err, domain.ErrorCodeCLAConfirmTokenInvalid) {
				t.Fatalf("err = %v, want cla_confirm_token_invalid", err)
			}

			if repo.saveCalled {
				t.Fatal("SaveNewCLA must not be called for an invalid token")
			}
		})
	}
}

func TestConfirmByTokenRepoErrorPassesThrough(t *testing.T) {
	dbErr := errors.New("db down")

	s := newServiceForConfirm(
		&fakeCLAConfirmToken{
			payload: repository.CLAConfirmTokenPayload{LinkId: "link1", Email: "alice@example.com"},
		},
		&fakeIndividualSigningRepo{findErr: dbErr},
		&fakeCLAService{claId: "9"},
		&fakeLinkRepo{},
	)

	if _, err := s.ConfirmByToken(&CmdToConfirmCLAByToken{Token: "t"}); !errors.Is(err, dbErr) {
		t.Fatalf("err = %v, want the raw db error", err)
	}
}

func TestConfirmByTokenRedisErrorPassesThrough(t *testing.T) {
	redisErr := errors.New("redis down")

	s := newServiceForConfirm(
		&fakeCLAConfirmToken{err: redisErr},
		&fakeIndividualSigningRepo{},
		&fakeCLAService{claId: "9"},
		&fakeLinkRepo{},
	)

	if _, err := s.ConfirmByToken(&CmdToConfirmCLAByToken{Token: "t"}); !errors.Is(err, redisErr) {
		t.Fatalf("err = %v, want the raw redis error", err)
	}
}

func TestConfirmByTokenCLANotExists(t *testing.T) {
	s := newServiceForConfirm(
		&fakeCLAConfirmToken{
			payload: repository.CLAConfirmTokenPayload{LinkId: "link1", Email: "alice@example.com"},
		},
		&fakeIndividualSigningRepo{sign: signedIndividual("7")},
		&fakeCLAService{claId: ""}, // no latest cla
		&fakeLinkRepo{},
	)

	_, err := s.ConfirmByToken(&CmdToConfirmCLAByToken{Token: "t"})
	if !domain.IsErrorOf(err, domain.ErrorCodeCLANotExists) {
		t.Fatalf("err = %v, want cla_not_exists", err)
	}
}

func TestConfirmByTokenAlreadyLatest(t *testing.T) {
	s := newServiceForConfirm(
		&fakeCLAConfirmToken{
			payload: repository.CLAConfirmTokenPayload{LinkId: "link1", Email: "alice@example.com"},
		},
		&fakeIndividualSigningRepo{sign: signedIndividual("9")},
		&fakeCLAService{claId: "9"}, // signing already on the latest cla
		&fakeLinkRepo{},
	)

	_, err := s.ConfirmByToken(&CmdToConfirmCLAByToken{Token: "t"})
	if !domain.IsErrorOf(err, domain.ErrorCodeIndividualSigningCLAIsLatest) {
		t.Fatalf("err = %v, want individual_signing_cla_is_latest", err)
	}
}

func TestConfirmByTokenSaveError(t *testing.T) {
	saveErr := errors.New("save failed")
	repo := &fakeIndividualSigningRepo{sign: signedIndividual("7"), saveErr: saveErr}

	s := newServiceForConfirm(
		&fakeCLAConfirmToken{
			payload: repository.CLAConfirmTokenPayload{LinkId: "link1", Email: "alice@example.com"},
		},
		repo,
		&fakeCLAService{claId: "9"},
		&fakeLinkRepo{},
	)

	if _, err := s.ConfirmByToken(&CmdToConfirmCLAByToken{Token: "t"}); !errors.Is(err, saveErr) {
		t.Fatalf("err = %v, want the save error", err)
	}
}

func TestConfirmByTokenSuccessWithoutOrgAlias(t *testing.T) {
	s := newServiceForConfirm(
		&fakeCLAConfirmToken{
			payload: repository.CLAConfirmTokenPayload{LinkId: "link1", Email: "alice@example.com"},
		},
		&fakeIndividualSigningRepo{sign: signedIndividual("7")},
		&fakeCLAService{claId: "9"},
		&fakeLinkRepo{findErr: errors.New("link not found")}, // degraded: no org alias
	)

	dto, err := s.ConfirmByToken(&CmdToConfirmCLAByToken{Token: "t"})
	if err != nil {
		t.Fatalf("ConfirmByToken should still succeed, got: %v", err)
	}
	if dto.Result != "confirmed" || dto.OrgAlias != "" {
		t.Fatalf("unexpected dto: %+v", dto)
	}
}

func TestCLAConfirmResultDTOAuditFieldsNotSerialized(t *testing.T) {
	// LinkId/ClaId are audit-only; keep them out of the json response contract.
	dto := CLAConfirmResultDTO{
		Result: "confirmed", OrgAlias: "o", EmailMasked: "a***@x.com",
		LinkId: "secret-link", ClaId: "9",
	}

	b, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	got := string(b)
	if want := `"org_alias":"o"`; !strings.Contains(got, want) {
		t.Fatalf("response %s should contain %s", got, want)
	}
	if strings.Contains(got, "secret-link") || strings.Contains(got, `"claid"`) || strings.Contains(got, `"link_id"`) {
		t.Fatalf("response %s must not contain audit fields", got)
	}
}

func TestToCLAConfirmTokenInvalid(t *testing.T) {
	if err := toCLAConfirmTokenInvalid(errors.New("x")); err.Error() != "x" {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := toCLAConfirmTokenInvalid(nil); err != nil {
		t.Fatalf("nil should stay nil, got: %v", err)
	}

	err := toCLAConfirmTokenInvalid(commonRepo.NewErrorResourceNotFound(errors.New("gone")))
	if !domain.IsErrorOf(err, domain.ErrorCodeCLAConfirmTokenInvalid) {
		t.Fatalf("unexpected error: %v", err)
	}
}
