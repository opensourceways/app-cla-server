package watch

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/worker"
)

func newTestImpl() *notifyAdminWatchImpl {
	return &notifyAdminWatchImpl{
		config: &NotifyAdminConfig{
			NotifyCorpAdminRemindDays:  90,
			NotifyIndividualRemindDays: 90,
			NotifyBatchSize:            500,
		},
		defaultGracePeriodDays: 30,
	}
}

func TestHandleCorpSigningNoPDF(t *testing.T) {
	impl := newTestImpl()

	link := &repository.LinkCLA{Id: "link1"}
	corp := &repository.CorpSigningSummary{HasPDF: false}

	impl.handleCorpSigning(link, corp)
}

func TestHandleCorpSigningAlreadyLatest(t *testing.T) {
	impl := newTestImpl()

	link := &repository.LinkCLA{
		Id: "link1",
		Clas: []domain.CLA{
			{Id: "10", Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")},
		},
	}
	corp := &repository.CorpSigningSummary{
		HasPDF: true,
		Link:   domain.LinkInfo{CLAInfo: domain.CLAInfo{CLAId: "10", Language: dp.CreateLanguage("en")}},
	}

	impl.handleCorpSigning(link, corp)
}

func TestHandleCorpSigningNoLatestCorpCLA(t *testing.T) {
	impl := newTestImpl()

	link := &repository.LinkCLA{
		Id:   "link1",
		Clas: []domain.CLA{},
	}
	corp := &repository.CorpSigningSummary{
		HasPDF: true,
		Link:   domain.LinkInfo{CLAInfo: domain.CLAInfo{CLAId: "9", Language: dp.CreateLanguage("en")}},
	}

	impl.handleCorpSigning(link, corp)
}

func TestHandleCorpSigningNotifyWithin7Days(t *testing.T) {
	impl := newTestImpl()

	link := &repository.LinkCLA{
		Id: "link1",
		Clas: []domain.CLA{
			{Id: "10", Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")},
		},
	}

	corp := &repository.CorpSigningSummary{
		HasPDF:    true,
		CLANotify: "10",
		Link:      domain.LinkInfo{CLAInfo: domain.CLAInfo{CLAId: "9", Language: dp.CreateLanguage("en")}},
	}

	impl.handleCorpSigning(link, corp)
}

func TestHandleIndividualSigningAlreadyLatest(t *testing.T) {
	impl := newTestImpl()

	link := &repository.LinkCLA{
		Id: "link1",
		Clas: []domain.CLA{
			{Id: "10", Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("en")},
		},
	}

	is := &domain.IndividualSigning{
		Link: domain.LinkInfo{CLAInfo: domain.CLAInfo{CLAId: "10", Language: dp.CreateLanguage("en")}},
	}

	impl.handleIndividualSigning(link, is)
}

func TestHandleIndividualSigningNoLatestCLA(t *testing.T) {
	impl := newTestImpl()

	link := &repository.LinkCLA{
		Id:   "link1",
		Clas: []domain.CLA{},
	}

	is := &domain.IndividualSigning{
		Link: domain.LinkInfo{CLAInfo: domain.CLAInfo{CLAId: "9", Language: dp.CreateLanguage("en")}},
	}

	impl.handleIndividualSigning(link, is)
}

// fakeEmailWorker 用于在测试中替换 worker.GetEmailWorker()，记录发信调用。
type fakeEmailWorker struct {
	sendCount int
	platform  string
	to        []string
	subject   string
	content   string
}

func (f *fakeEmailWorker) SendSimpleMessage(platform string, msg *worker.EmailMessage) {
	f.sendCount++
	f.platform = platform
	if msg != nil {
		f.to = msg.To
		f.subject = msg.Subject
		f.content = msg.Content.String()
	}
}

func (f *fakeEmailWorker) GenCLAPDFForCorporationAndSendIt(string, *models.OrgInfo, *models.CLAInfo, *models.CorporationSigning) {
}

func (f *fakeEmailWorker) Shutdown() {}

// TestHandleSendEmailAuditLogCorp 注入假 worker 并直接调用 handleSendEmail，
// 验证企业 CLA 变更通知的发信（含 [audit] cla_notify_email: scene=corp 审计日志）路径被走到、不 panic。
func TestHandleSendEmailAuditLogCorp(t *testing.T) {
	fw := &fakeEmailWorker{}
	restore := worker.SetEmailWorker(fw)
	defer restore()

	impl := newTestImpl()

	link := &repository.LinkCLA{
		Id:  "link1",
		Org: domain.OrgInfo{Alias: "test-org", ProjectURL: "http://project.example.com"},
		Email: domain.EmailInfo{
			Addr:     dp.CreateEmailAddr("from@example.com"),
			Platform: "plat1",
		},
	}
	corp := &repository.CorpSigningSummary{
		Id: "corp1",
		Admin: domain.Manager{Representative: domain.Representative{
			Name:      dp.CreateName("admin"),
			EmailAddr: dp.CreateEmailAddr("admin@example.com"),
		}},
		CLANotify:      "new-cla-id",
		Link:           domain.LinkInfo{CLAInfo: domain.CLAInfo{CLAId: "old-cla-id"}},
		ClaNotifyCount: 2,
	}

	if err := impl.handleSendEmail(link, corp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fw.sendCount != 1 {
		t.Fatalf("expected SendSimpleMessage called once, got %d", fw.sendCount)
	}
	if fw.subject != "CLA has been updated" {
		t.Fatalf("unexpected subject: %s", fw.subject)
	}
	if len(fw.to) != 1 || fw.to[0] != "admin@example.com" {
		t.Fatalf("unexpected recipient: %v", fw.to)
	}
}

// TestHandleSendIndividualEmailAuditLogIndividual 注入假 worker 并直接调用 handleSendIndividualEmail，
// 验证个人 CLA 变更通知的发信（含 [audit] cla_notify_email: scene=individual 审计日志）路径被走到、不 panic。
func TestHandleSendIndividualEmailAuditLogIndividual(t *testing.T) {
	fw := &fakeEmailWorker{}
	restore := worker.SetEmailWorker(fw)
	defer restore()

	impl := newTestImpl()
	impl.claPlatformURL = "http://cla.example.com/sign/"

	link := &repository.LinkCLA{
		Id:  "link1",
		Org: domain.OrgInfo{Alias: "test-org", ProjectURL: "http://project.example.com"},
		Email: domain.EmailInfo{
			Addr:     dp.CreateEmailAddr("from@example.com"),
			Platform: "plat1",
		},
	}
	is := &domain.IndividualSigning{
		Rep: domain.Representative{
			Name:      dp.CreateName("alice"),
			EmailAddr: dp.CreateEmailAddr("alice@example.com"),
		},
		Link:           domain.LinkInfo{CLAInfo: domain.CLAInfo{CLAId: "old-cla-id"}},
		ClaNotifyCount: 3,
	}
	latestCLA := &domain.CLA{Id: "new-cla-id", UpdatedAt: time.Now().Unix()}

	if err := impl.handleSendIndividualEmail(link, is, latestCLA); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fw.sendCount != 1 {
		t.Fatalf("expected SendSimpleMessage called once, got %d", fw.sendCount)
	}
	if fw.subject != "CLA has been updated - Action Required" {
		t.Fatalf("unexpected subject: %s", fw.subject)
	}
	if len(fw.to) != 1 || fw.to[0] != "alice@example.com" {
		t.Fatalf("unexpected recipient: %v", fw.to)
	}
}

// fakeCLAConfirmTokenRepo 记录 Add 调用并返回可注入的结果。
type fakeCLAConfirmTokenRepo struct {
	addedLinkIds []string
	addedEmails  []string
	addedClaIds  []string
	addedTTLs    []time.Duration
	token        string
	err          error
}

func (f *fakeCLAConfirmTokenRepo) Add(linkId, email, newClaId string, ttl time.Duration) (string, error) {
	f.addedLinkIds = append(f.addedLinkIds, linkId)
	f.addedEmails = append(f.addedEmails, email)
	f.addedClaIds = append(f.addedClaIds, newClaId)
	f.addedTTLs = append(f.addedTTLs, ttl)

	if f.err != nil {
		return "", f.err
	}

	return f.token, nil
}

func (f *fakeCLAConfirmTokenRepo) Consume(token string) (repository.CLAConfirmTokenPayload, error) {
	return repository.CLAConfirmTokenPayload{}, nil
}

func newIndividualEmailCase() (*notifyAdminWatchImpl, *repository.LinkCLA, *domain.IndividualSigning, *domain.CLA) {
	impl := newTestImpl()
	impl.claPlatformURL = "http://cla.example.com/sign/"
	impl.config.SendEmailInterval = 0

	link := &repository.LinkCLA{
		Id:  "link1",
		Org: domain.OrgInfo{Alias: "test-org", ProjectURL: "http://project.example.com"},
		Email: domain.EmailInfo{
			Addr:     dp.CreateEmailAddr("from@example.com"),
			Platform: "plat1",
		},
	}
	is := &domain.IndividualSigning{
		Rep: domain.Representative{
			Name:      dp.CreateName("alice"),
			EmailAddr: dp.CreateEmailAddr("alice@example.com"),
		},
		Link:           domain.LinkInfo{CLAInfo: domain.CLAInfo{CLAId: "old-cla-id"}},
		ClaNotifyCount: 3,
	}
	latestCLA := &domain.CLA{Id: "new-cla-id", UpdatedAt: time.Now().Unix()}

	return impl, link, is, latestCLA
}

// TestHandleSendIndividualEmailWithConfirmToken 令牌生成成功时，邮件正文应包含
// 一键确认链接（含 token 与有效期）与原验证码签署链接。
func TestHandleSendIndividualEmailWithConfirmToken(t *testing.T) {
	fw := &fakeEmailWorker{}
	restore := worker.SetEmailWorker(fw)
	defer restore()

	impl, link, is, latestCLA := newIndividualEmailCase()
	impl.claConfirmToken = &fakeCLAConfirmTokenRepo{token: "a1b2c3d4e5f6a7b8a1b2c3d4e5f6a7b8a1b2c3d4e5f6a7b8a1b2c3d4e5f6a7b8"}

	if err := impl.handleSendIndividualEmail(link, is, latestCLA); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fw.sendCount != 1 {
		t.Fatalf("expected one email, got %d", fw.sendCount)
	}

	wantConfirm := "/confirm-cla/link1?t=a1b2c3d4e5f6a7b8"
	if !strings.Contains(fw.content, wantConfirm) {
		t.Fatalf("email content should contain confirm url %q, content:\n%s", wantConfirm, fw.content)
	}
	if !strings.Contains(fw.content, "one-time use") {
		t.Fatalf("email content should mention one-time use, content:\n%s", fw.content)
	}
	if !strings.Contains(fw.content, "/sign-cla/link1/individual-update?email=alice@example.com") {
		t.Fatalf("email content should still contain the verification-code url, content:\n%s", fw.content)
	}
	if !strings.Contains(fw.content, "verification code") {
		t.Fatalf("email content should mention the fallback, content:\n%s", fw.content)
	}
}

// TestHandleSendIndividualEmailTokenGenFailed 令牌生成失败时降级：邮件仍发送，
// 正文不含一键确认链接，只保留验证码签署链接。
func TestHandleSendIndividualEmailTokenGenFailed(t *testing.T) {
	fw := &fakeEmailWorker{}
	restore := worker.SetEmailWorker(fw)
	defer restore()

	impl, link, is, latestCLA := newIndividualEmailCase()
	impl.claConfirmToken = &fakeCLAConfirmTokenRepo{err: errors.New("redis down")}

	if err := impl.handleSendIndividualEmail(link, is, latestCLA); err != nil {
		t.Fatalf("degraded email sending should still succeed, got: %v", err)
	}

	if fw.sendCount != 1 {
		t.Fatalf("email must still be sent when token generation fails, got %d", fw.sendCount)
	}
	if strings.Contains(fw.content, "confirm-cla") {
		t.Fatalf("degraded email should not contain the confirm url, content:\n%s", fw.content)
	}
	if !strings.Contains(fw.content, "/sign-cla/link1/individual-update?email=alice@example.com") {
		t.Fatalf("degraded email must contain the verification-code url, content:\n%s", fw.content)
	}
}

// TestGenCLAConfirmURL 校验令牌生成入参（link_id/email/new_cla_id/ttl）与链接拼装。
func TestGenCLAConfirmURL(t *testing.T) {
	impl, link, is, latestCLA := newIndividualEmailCase()
	impl.config.CLAConfirmTokenExpiry = 604800
	fake := &fakeCLAConfirmTokenRepo{token: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"}
	impl.claConfirmToken = fake

	url, days := impl.genCLAConfirmURL(link, is, latestCLA)
	if url != "http://cla.example.com/confirm-cla/link1?t=ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" {
		t.Fatalf("unexpected confirm url: %s", url)
	}
	if days != 7 {
		t.Fatalf("valid days = %d, want 7", days)
	}
	if len(fake.addedLinkIds) != 1 || fake.addedLinkIds[0] != "link1" {
		t.Fatalf("unexpected link ids: %v", fake.addedLinkIds)
	}
	if len(fake.addedEmails) != 1 || fake.addedEmails[0] != "alice@example.com" {
		t.Fatalf("unexpected emails: %v", fake.addedEmails)
	}
	if len(fake.addedClaIds) != 1 || fake.addedClaIds[0] != "new-cla-id" {
		t.Fatalf("unexpected cla ids: %v", fake.addedClaIds)
	}
	if len(fake.addedTTLs) != 1 || fake.addedTTLs[0] != 604800*time.Second {
		t.Fatalf("unexpected ttls: %v", fake.addedTTLs)
	}
}

// TestGenCLAConfirmURLNilRepo 未注入令牌仓储（nil）时应降级为空链接，不 panic。
func TestGenCLAConfirmURLNilRepo(t *testing.T) {
	impl, link, is, latestCLA := newIndividualEmailCase()
	impl.claConfirmToken = nil

	url, days := impl.genCLAConfirmURL(link, is, latestCLA)
	if url != "" || days != 0 {
		t.Fatalf("nil repo should degrade to empty url, got (%q, %d)", url, days)
	}
}
