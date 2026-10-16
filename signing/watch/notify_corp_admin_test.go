package watch

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
	"github.com/opensourceways/app-cla-server/worker"
	"github.com/opensourceways/obs-sdk/go/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
}

func (f *fakeEmailWorker) SendSimpleMessage(platform string, msg *worker.EmailMessage) {
	f.sendCount++
	f.platform = platform
	if msg != nil {
		f.to = msg.To
		f.subject = msg.Subject
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

// initWatchTestLogger replaces the global default logger with one writing
// to a buffer, returning the buffer and a cleanup to restore the original.
func initWatchTestLogger(t *testing.T) (*bytes.Buffer, func()) {
	t.Helper()
	var buf bytes.Buffer
	orig := log.Default()
	log.SetDefault(log.New(log.Config{
		Service:   "app-cla-server",
		Env:       "test",
		Instance:  "pod-1",
		Community: "openeuler",
		Output:    &buf,
	}))
	return &buf, func() { log.SetDefault(orig) }
}

func parseWatchLogLines(t *testing.T, buf *bytes.Buffer) []map[string]interface{} {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	result := make([]map[string]interface{}, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var m map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(line), &m), "line: %s", line)
		result = append(result, m)
	}
	return result
}

func findWatchLogEntryByMsg(t *testing.T, entries []map[string]interface{}, msg string) map[string]interface{} {
	t.Helper()
	for _, e := range entries {
		if e["msg"] == msg {
			return e
		}
	}
	t.Fatalf("log entry with msg=%q not found in %d entries", msg, len(entries))
	return nil
}

// TestAuditLogMsgClaNotifyEmailCorp verifies that the JSON log line emitted
// on corp CLA change notification has msg="[audit] cla_notify_email"
// (underscore event identifier, not space-separated) with scene="corp".
func TestAuditLogMsgClaNotifyEmailCorp(t *testing.T) {
	buf, cleanup := initWatchTestLogger(t)
	defer cleanup()

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

	err := impl.handleSendEmail(link, corp)
	require.NoError(t, err)

	entries := parseWatchLogLines(t, buf)
	e := findWatchLogEntryByMsg(t, entries, "[audit] cla_notify_email")

	assert.Equal(t, "info", e["level"])
	assert.Equal(t, "corp", e["scene"])
	assert.Equal(t, "link1", e["link_id"])
}

// TestAuditLogMsgClaNotifyEmailIndividual verifies that the JSON log line
// emitted on individual CLA change notification has msg="[audit] cla_notify_email"
// (underscore event identifier) with scene="individual".
func TestAuditLogMsgClaNotifyEmailIndividual(t *testing.T) {
	buf, cleanup := initWatchTestLogger(t)
	defer cleanup()

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

	err := impl.handleSendIndividualEmail(link, is, latestCLA)
	require.NoError(t, err)

	entries := parseWatchLogLines(t, buf)
	e := findWatchLogEntryByMsg(t, entries, "[audit] cla_notify_email")

	assert.Equal(t, "info", e["level"])
	assert.Equal(t, "individual", e["scene"])
	assert.Equal(t, "link1", e["link_id"])
}
