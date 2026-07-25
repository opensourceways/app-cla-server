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

// ---- mocks ----

type mockUserService struct {
	userservice.UserService // embed nil interface; only overridden methods are used

	validUser      bool
	validErr       error
	updateEmailErr error
	updateByAcctErr error
	addPws         map[string]dp.Password
	addIDs         []string
	addErr         error

	calledIsAValid bool
	calledUpdateEmail bool
	calledUpdateByAcct bool
	calledAdd         bool

	isValidEmailArg  dp.EmailAddr
	updateEmailOld    dp.EmailAddr
	updateEmailNew    dp.EmailAddr
	updateByAcctAcct  dp.Account
	updateByAcctEmail dp.EmailAddr
	addManagers       []domain.Manager
}

func (m *mockUserService) IsAValidUser(linkId string, email dp.EmailAddr) (bool, error) {
	m.calledIsAValid = true
	m.isValidEmailArg = email
	return m.validUser, m.validErr
}

func (m *mockUserService) UpdateEmail(linkId string, oldEmail, newEmail dp.EmailAddr) error {
	m.calledUpdateEmail = true
	m.updateEmailOld = oldEmail
	m.updateEmailNew = newEmail
	return m.updateEmailErr
}

func (m *mockUserService) UpdateEmailByAccount(linkId string, account dp.Account, newEmail dp.EmailAddr) error {
	m.calledUpdateByAcct = true
	m.updateByAcctAcct = account
	m.updateByAcctEmail = newEmail
	return m.updateByAcctErr
}

func (m *mockUserService) Add(linkId, csId string, managers []domain.Manager) (map[string]dp.Password, []string, error) {
	m.calledAdd = true
	m.addManagers = managers
	return m.addPws, m.addIDs, m.addErr
}

type mockLinkRepo struct {
	repository.Link
	link domain.Link
	err  error
}

func (m *mockLinkRepo) Find(string) (domain.Link, error) {
	return m.link, m.err
}

type mockCorpSigningRepo struct {
	repository.CorpSigning
	findResult domain.CorpSigning
	findErr    error
	updateErr  error
	updated    *domain.CorpSigning
}

func (m *mockCorpSigningRepo) Find(string) (domain.CorpSigning, error) {
	return m.findResult, m.findErr
}

func (m *mockCorpSigningRepo) Update(cs *domain.CorpSigning) error {
	m.updated = cs
	return m.updateErr
}

// ---- helpers ----

func newService(repo repository.CorpSigning, linkRepo repository.Link, us userservice.UserService) *corpSigningService {
	return &corpSigningService{repo: repo, linkRepo: linkRepo, userService: us}
}

func mustName(s string) dp.Name {
	v, err := dp.NewName(s)
	if err != nil {
		panic(err)
	}
	return v
}

func mustEmail(s string) dp.EmailAddr {
	v, err := dp.NewEmailAddr(s)
	if err != nil {
		panic(err)
	}
	return v
}

func mustPW(s string) dp.Password {
	v, err := dp.NewPassword([]byte(s))
	if err != nil {
		panic(err)
	}
	return v
}

// cs with admin set, rep email == admin email == a@corp.com
func baseCorpSigning() domain.CorpSigning {
	rep := domain.Representative{Name: mustName("Alice"), EmailAddr: mustEmail("a@corp.com")}
	return domain.CorpSigning{
		Id:   "cs1",
		Link: domain.LinkInfo{Id: "link1"},
		Rep:  rep,
		Admin: domain.Manager{
			Id:             "admin",
			Representative: rep,
		},
	}
}

// ---- tests ----

// 用例1: 账号已存在(同域换人) -> UpdateEmail, 不新建, dto=nil
func TestUpdateRepresentative_AccountExists_SameDomain(t *testing.T) {
	cs := baseCorpSigning()
	repo := &mockCorpSigningRepo{findResult: cs}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	us := &mockUserService{validUser: true}

	s := newService(repo, linkRepo, us)
	dto, err := s.UpdateRepresentative("user1", "link1", "cs1", "Bob", "b@corp.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto != nil {
		t.Errorf("dto should be nil when reusing existing account, got %+v", dto)
	}
	if !us.calledIsAValid {
		t.Error("IsAValidUser should be called")
	}
	if !us.calledUpdateEmail {
		t.Error("UpdateEmail should be called")
	}
	if us.calledAdd {
		t.Error("Add should NOT be called when account exists")
	}
	if us.updateEmailOld.EmailAddr() != "a@corp.com" {
		t.Errorf("old email: got %s", us.updateEmailOld.EmailAddr())
	}
	if us.updateEmailNew.EmailAddr() != "b@corp.com" {
		t.Errorf("new email: got %s", us.updateEmailNew.EmailAddr())
	}
}

// 用例2: 账号已存在(跨域换人) -> UpdateEmail, account 由 userservice 内部改写
func TestUpdateRepresentative_AccountExists_CrossDomain(t *testing.T) {
	cs := baseCorpSigning()
	repo := &mockCorpSigningRepo{findResult: cs}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	us := &mockUserService{validUser: true}

	s := newService(repo, linkRepo, us)
	dto, err := s.UpdateRepresentative("user1", "link1", "cs1", "Bob", "b@newcorp.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto != nil {
		t.Errorf("dto should be nil, got %+v", dto)
	}
	if !us.calledUpdateEmail {
		t.Error("UpdateEmail should be called")
	}
	if us.updateEmailNew.EmailAddr() != "b@newcorp.com" {
		t.Errorf("new email: got %s", us.updateEmailNew.EmailAddr())
	}
}

// 用例3(核心): 账号未预置 -> Add 新建, dto!=nil 含初始口令
func TestUpdateRepresentative_AccountMissing_CreateNew(t *testing.T) {
	cs := baseCorpSigning()
	repo := &mockCorpSigningRepo{findResult: cs}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	us := &mockUserService{
		validUser: false,
		addPws:    map[string]dp.Password{"admin": mustPW("init-pw")},
		addIDs:    []string{"u1"},
	}

	s := newService(repo, linkRepo, us)
	dto, err := s.UpdateRepresentative("user1", "link1", "cs1", "Bob", "b@corp.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto == nil {
		t.Fatal("dto should not be nil when a new account is created")
	}
	if !us.calledAdd {
		t.Error("Add should be called when account is missing")
	}
	if us.calledUpdateEmail || us.calledUpdateByAcct {
		t.Error("UpdateEmail/UpdateEmailByAccount should NOT be called")
	}
	if dto.Role != domain.RoleAdmin {
		t.Errorf("role: got %s, want %s", dto.Role, domain.RoleAdmin)
	}
	if dto.Name != "Bob" {
		t.Errorf("name: got %s, want Bob", dto.Name)
	}
	if dto.EmailAddr != "b@corp.com" {
		t.Errorf("email: got %s, want b@corp.com", dto.EmailAddr)
	}
	if dto.Account != "admin_corp.com" {
		t.Errorf("account: got %s, want admin_corp.com", dto.Account)
	}
	if string(dto.Password) != "init-pw" {
		t.Errorf("password: got %s", string(dto.Password))
	}
	if len(us.addManagers) != 1 || us.addManagers[0].Id != "admin" {
		t.Errorf("add managers: %+v", us.addManagers)
	}
}

// 用例4: 邮箱未变 -> 跳过账号同步
func TestUpdateRepresentative_EmailUnchanged(t *testing.T) {
	cs := baseCorpSigning()
	repo := &mockCorpSigningRepo{findResult: cs}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	us := &mockUserService{validUser: true}

	s := newService(repo, linkRepo, us)
	dto, err := s.UpdateRepresentative("user1", "link1", "cs1", "NewName", "a@corp.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto != nil {
		t.Errorf("dto should be nil, got %+v", dto)
	}
	if us.calledIsAValid || us.calledUpdateEmail || us.calledAdd {
		t.Error("no user service call expected when email is unchanged")
	}
}

// 用例5: 管理员尚未创建(cs.Admin.Id=="") -> 仅更新 rep, 不碰 user 表
func TestUpdateRepresentative_AdminNotCreated(t *testing.T) {
	cs := baseCorpSigning()
	cs.Admin.Id = ""
	repo := &mockCorpSigningRepo{findResult: cs}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	us := &mockUserService{}

	s := newService(repo, linkRepo, us)
	dto, err := s.UpdateRepresentative("user1", "link1", "cs1", "Bob", "b@corp.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto != nil {
		t.Errorf("dto should be nil, got %+v", dto)
	}
	if us.calledIsAValid || us.calledUpdateEmail || us.calledAdd || us.calledUpdateByAcct {
		t.Error("no user service call expected when admin id is empty")
	}
}

// 用例6(异常): 孤儿账号 -> Add 返回 user_exists -> 按账号回退 UpdateEmailByAccount
func TestUpdateRepresentative_OrphanAccount_FallbackByAccount(t *testing.T) {
	cs := baseCorpSigning()
	repo := &mockCorpSigningRepo{findResult: cs}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	us := &mockUserService{
		validUser: false,
		addErr:    domain.NewDomainError(domain.ErrorCodeUserExists),
	}

	s := newService(repo, linkRepo, us)
	dto, err := s.UpdateRepresentative("user1", "link1", "cs1", "Bob", "b@corp.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dto != nil {
		t.Errorf("dto should be nil for orphan fallback, got %+v", dto)
	}
	if !us.calledAdd {
		t.Error("Add should be attempted first")
	}
	if !us.calledUpdateByAcct {
		t.Error("UpdateEmailByAccount should be called as fallback")
	}
	if us.updateByAcctAcct.Account() != "admin_corp.com" {
		t.Errorf("account: got %s, want admin_corp.com", us.updateByAcctAcct.Account())
	}
	if us.updateByAcctEmail.EmailAddr() != "b@corp.com" {
		t.Errorf("email: got %s", us.updateByAcctEmail.EmailAddr())
	}
}

// 用例7(异常): Add 返回非 user_exists 错误 -> 直接返回错误, 不回退
func TestUpdateRepresentative_AddOtherError(t *testing.T) {
	cs := baseCorpSigning()
	repo := &mockCorpSigningRepo{findResult: cs}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	wantErr := errors.New("db down")
	us := &mockUserService{
		validUser: false,
		addErr:    wantErr,
	}

	s := newService(repo, linkRepo, us)
	dto, err := s.UpdateRepresentative("user1", "link1", "cs1", "Bob", "b@corp.com")

	if err != wantErr {
		t.Fatalf("err: got %v, want %v", err, wantErr)
	}
	if dto != nil {
		t.Errorf("dto should be nil, got %+v", dto)
	}
	if us.calledUpdateByAcct {
		t.Error("UpdateEmailByAccount should NOT be called for non-user_exists error")
	}
}

// 用例8(非法): 权限不足 -> 返回错误, 不读 corp_signing/不动 user 表
func TestUpdateRepresentative_NoPermission(t *testing.T) {
	cs := baseCorpSigning()
	repo := &mockCorpSigningRepo{findResult: cs}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "owner"}}
	us := &mockUserService{}

	s := newService(repo, linkRepo, us)
	_, err := s.UpdateRepresentative("stranger", "link1", "cs1", "Bob", "b@corp.com")

	if err == nil {
		t.Fatal("expected permission error")
	}
	if us.calledIsAValid || us.calledAdd || us.calledUpdateEmail {
		t.Error("no user service call expected when permission denied")
	}
	if repo.updated != nil {
		t.Error("repo.Update should NOT be called when permission denied")
	}
}

// 用例9(非法): link_id 不匹配 -> 返回 not found
func TestUpdateRepresentative_LinkIdMismatch(t *testing.T) {
	cs := baseCorpSigning()
	repo := &mockCorpSigningRepo{findResult: cs}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	us := &mockUserService{}

	s := newService(repo, linkRepo, us)
	_, err := s.UpdateRepresentative("user1", "other-link", "cs1", "Bob", "b@corp.com")

	if err == nil {
		t.Fatal("expected error for link_id mismatch")
	}
	if !commonRepo.IsErrorResourceNotFound(err) {
		t.Errorf("err should be ErrorResourceNotFound, got %T", err)
	}
	if us.calledIsAValid || us.calledAdd {
		t.Error("no user service call expected on link mismatch")
	}
}

// 用例10(异常): repo.Find 失败
func TestUpdateRepresentative_FindError(t *testing.T) {
	findErr := errors.New("find failed")
	repo := &mockCorpSigningRepo{findErr: findErr}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	us := &mockUserService{}

	s := newService(repo, linkRepo, us)
	_, err := s.UpdateRepresentative("user1", "link1", "cs1", "Bob", "b@corp.com")

	if err != findErr {
		t.Fatalf("err: got %v, want %v", err, findErr)
	}
}

// 用例11(异常): IsAValidUser 失败
func TestUpdateRepresentative_IsAValidUserError(t *testing.T) {
	cs := baseCorpSigning()
	repo := &mockCorpSigningRepo{findResult: cs}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	validErr := errors.New("db error")
	us := &mockUserService{validErr: validErr}

	s := newService(repo, linkRepo, us)
	_, err := s.UpdateRepresentative("user1", "link1", "cs1", "Bob", "b@corp.com")

	if err != validErr {
		t.Fatalf("err: got %v, want %v", err, validErr)
	}
	if us.calledUpdateEmail || us.calledAdd {
		t.Error("no further user service call expected after IsAValidUser error")
	}
}

// 用例12(异常): UpdateEmail 失败
func TestUpdateRepresentative_UpdateEmailError(t *testing.T) {
	cs := baseCorpSigning()
	repo := &mockCorpSigningRepo{findResult: cs}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	updateErr := errors.New("update failed")
	us := &mockUserService{validUser: true, updateEmailErr: updateErr}

	s := newService(repo, linkRepo, us)
	_, err := s.UpdateRepresentative("user1", "link1", "cs1", "Bob", "b@corp.com")

	if err != updateErr {
		t.Fatalf("err: got %v, want %v", err, updateErr)
	}
}

// 用例13(异常): 孤儿回退 UpdateEmailByAccount 失败
func TestUpdateRepresentative_FallbackByAccountError(t *testing.T) {
	cs := baseCorpSigning()
	repo := &mockCorpSigningRepo{findResult: cs}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	acctErr := errors.New("find by account failed")
	us := &mockUserService{
		validUser:       false,
		addErr:          domain.NewDomainError(domain.ErrorCodeUserExists),
		updateByAcctErr: acctErr,
	}

	s := newService(repo, linkRepo, us)
	_, err := s.UpdateRepresentative("user1", "link1", "cs1", "Bob", "b@corp.com")

	if err != acctErr {
		t.Fatalf("err: got %v, want %v", err, acctErr)
	}
}

// 用例14(异常): rep 参数非法(空邮箱) -> 返回错误
func TestUpdateRepresentative_InvalidEmail(t *testing.T) {
	repo := &mockCorpSigningRepo{findResult: baseCorpSigning()}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	us := &mockUserService{}

	s := newService(repo, linkRepo, us)
	_, err := s.UpdateRepresentative("user1", "link1", "cs1", "Bob", "not-an-email")

	if err == nil {
		t.Fatal("expected error for invalid email")
	}
	if repo.updated != nil {
		t.Error("repo.Update should NOT be called for invalid email")
	}
}

// 用例15(异常): repo.Update 失败
func TestUpdateRepresentative_UpdateError(t *testing.T) {
	cs := baseCorpSigning()
	updateErr := errors.New("update failed")
	repo := &mockCorpSigningRepo{findResult: cs, updateErr: updateErr}
	linkRepo := &mockLinkRepo{link: domain.Link{Id: "link1", Submitter: "user1"}}
	us := &mockUserService{}

	s := newService(repo, linkRepo, us)
	_, err := s.UpdateRepresentative("user1", "link1", "cs1", "Bob", "b@corp.com")

	if err != updateErr {
		t.Fatalf("err: got %v, want %v", err, updateErr)
	}
	if us.calledIsAValid || us.calledAdd {
		t.Error("no user service call expected when repo.Update fails")
	}
}
