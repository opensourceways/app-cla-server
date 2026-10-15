package app

import (
	"bytes"
	"encoding/csv"
	"errors"
	"testing"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func TestCorpListToCSVBOM(t *testing.T) {
	data := corpListToCSV(nil)
	if !bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		t.Error("CSV output should start with UTF-8 BOM")
	}
}

func TestCorpListToCSVEmpty(t *testing.T) {
	data := corpListToCSV(nil)
	body := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	r := csv.NewReader(bytes.NewReader(body))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record (header only), got %d", len(records))
	}

	expected := []string{"序号", "企业名称", "申请状态", "CLA语言", "申请时间"}
	for i, v := range expected {
		if records[0][i] != v {
			t.Errorf("header[%d]: got %s, want %s", i, records[0][i], v)
		}
	}
}

func TestCorpListToCSVNormal(t *testing.T) {
	rows := []CorpSigningDTO{
		{CorpName: "华为技术有限公司", Language: "zh", Date: "2026-08-20 10:23:45", HasAdminAdded: true},
		{CorpName: "某某科技公司", Language: "en", Date: "2026-08-25 16:02:11", HasAdminAdded: false},
	}

	data := corpListToCSV(rows)
	body := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	r := csv.NewReader(bytes.NewReader(body))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("expected 3 records (header + 2 rows), got %d", len(records))
	}

	if records[1][0] != "1" {
		t.Errorf("row 1 index: got %s, want 1", records[1][0])
	}
	if records[1][1] != "华为技术有限公司" {
		t.Errorf("row 1 corp name: got %s, want 华为技术有限公司", records[1][1])
	}
	if records[1][2] != "已完成" {
		t.Errorf("row 1 status: got %s, want 已完成", records[1][2])
	}
	if records[1][3] != "zh" {
		t.Errorf("row 1 language: got %s, want zh", records[1][3])
	}
	if records[1][4] != "2026-08-20 10:23:45" {
		t.Errorf("row 1 date: got %s, want 2026-08-20 10:23:45", records[1][4])
	}

	if records[2][0] != "2" {
		t.Errorf("row 2 index: got %s, want 2", records[2][0])
	}
	if records[2][2] != "未完成" {
		t.Errorf("row 2 status: got %s, want 未完成", records[2][2])
	}
}

func TestCorpListToCSVFormulaInjection(t *testing.T) {
	rows := []CorpSigningDTO{
		{CorpName: "=cmd|calc", HasAdminAdded: false},
		{CorpName: "+1+1", HasAdminAdded: false},
		{CorpName: "-1+1", HasAdminAdded: false},
		{CorpName: "@cmd", HasAdminAdded: false},
	}

	data := corpListToCSV(rows)
	body := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	r := csv.NewReader(bytes.NewReader(body))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	expected := []string{"'=cmd|calc", "'+1+1", "'-1+1", "'@cmd"}
	for i, want := range expected {
		if records[i+1][1] != want {
			t.Errorf("row %d corp name: got %s, want %s", i+1, records[i+1][1], want)
		}
	}
}

func TestCorpListToCSVSpecialChars(t *testing.T) {
	rows := []CorpSigningDTO{
		{CorpName: "公司,有逗号", Language: "zh", Date: "2026-08-20 10:23:45", HasAdminAdded: true},
		{CorpName: `公司"有引号`, Language: "zh", Date: "2026-08-21 11:00:00", HasAdminAdded: false},
	}

	data := corpListToCSV(rows)
	body := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	r := csv.NewReader(bytes.NewReader(body))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	if records[1][1] != "公司,有逗号" {
		t.Errorf("comma cell: got %s, want 公司,有逗号", records[1][1])
	}
	if records[2][1] != `公司"有引号` {
		t.Errorf("quote cell: got %s, want 公司\"有引号", records[2][1])
	}
}

func TestEscapeCSVCell(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"equals prefix", "=cmd", "'=cmd"},
		{"plus prefix", "+1", "'+1"},
		{"minus prefix", "-1", "'-1"},
		{"at prefix", "@cmd", "'@cmd"},
		{"normal text", "华为", "华为"},
		{"empty string", "", ""},
		{"number prefix", "123", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeCSVCell(tt.input); got != tt.want {
				t.Errorf("escapeCSVCell(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---- mock implementations for Export method tests ----

type mockLinkRepo struct {
	repository.Link
	link domain.Link
	err  error
}

func (m *mockLinkRepo) Find(linkId string) (domain.Link, error) {
	return m.link, m.err
}

type mockCorpSigningRepo struct {
	repository.CorpSigning
	pages map[int]repository.CorpSigningSummaryPage
	err   error
}

func (m *mockCorpSigningRepo) FindPage(linkId string, page, pageSize int, adminAdded bool, searchQuery string) (repository.CorpSigningSummaryPage, error) {
	if m.err != nil {
		return repository.CorpSigningSummaryPage{}, m.err
	}
	if p, ok := m.pages[page]; ok {
		return p, nil
	}
	return repository.CorpSigningSummaryPage{Total: 0}, nil
}

func newTestCorpSigningService(repo repository.CorpSigning, linkRepo repository.Link) *corpSigningService {
	return &corpSigningService{
		repo:     repo,
		linkRepo: linkRepo,
	}
}

func TestExportSuccess(t *testing.T) {
	linkRepo := &mockLinkRepo{
		link: domain.Link{Id: "link1", Submitter: "user1"},
	}

	repo := &mockCorpSigningRepo{
		pages: map[int]repository.CorpSigningSummaryPage{
			1: {
				Total: 2,
				Data: []repository.CorpSigningSummary{
					{
						Id:   "cs1",
						Date: "2026-08-20 10:23:45",
						Link: domain.LinkInfo{
							Id:      "link1",
							CLAInfo: domain.CLAInfo{CLAId: "cla1", Language: dp.CreateLanguage("zh")},
						},
						Rep: domain.Representative{
							Name:      dp.CreateName("张三"),
							EmailAddr: dp.CreateEmailAddr("zhangsan@huawei.com"),
						},
						Corp:  domain.Corporation{Name: dp.CreateCorpName("华为技术有限公司")},
						Admin: domain.Manager{Id: "admin1"},
					},
					{
						Id:   "cs2",
						Date: "2026-08-25 16:02:11",
						Link: domain.LinkInfo{
							Id:      "link1",
							CLAInfo: domain.CLAInfo{CLAId: "cla1", Language: dp.CreateLanguage("en")},
						},
						Rep: domain.Representative{
							Name:      dp.CreateName("李四"),
							EmailAddr: dp.CreateEmailAddr("lisi@tech.com"),
						},
						Corp:  domain.Corporation{Name: dp.CreateCorpName("某某科技公司")},
						Admin: domain.Manager{},
					},
				},
			},
		},
	}

	s := newTestCorpSigningService(repo, linkRepo)
	data, err := s.Export("user1", "link1", false, "")
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	if !bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}) {
		t.Fatal("CSV output should start with UTF-8 BOM")
	}

	body := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	r := csv.NewReader(bytes.NewReader(body))
	records, rErr := r.ReadAll()
	if rErr != nil {
		t.Fatalf("failed to read CSV: %v", rErr)
	}

	if len(records) != 3 {
		t.Fatalf("expected 3 records (header + 2 rows), got %d", len(records))
	}

	if records[1][1] != "华为技术有限公司" {
		t.Errorf("row 1 corp name: got %s", records[1][1])
	}
	if records[1][2] != "已完成" {
		t.Errorf("row 1 status: got %s, want 已完成", records[1][2])
	}
	if records[2][2] != "未完成" {
		t.Errorf("row 2 status: got %s, want 未完成", records[2][2])
	}
}

func TestExportEmptyResult(t *testing.T) {
	linkRepo := &mockLinkRepo{
		link: domain.Link{Id: "link1", Submitter: "user1"},
	}

	repo := &mockCorpSigningRepo{
		pages: map[int]repository.CorpSigningSummaryPage{},
	}

	s := newTestCorpSigningService(repo, linkRepo)
	data, err := s.Export("user1", "link1", true, "")
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	body := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	r := csv.NewReader(bytes.NewReader(body))
	records, rErr := r.ReadAll()
	if rErr != nil {
		t.Fatalf("failed to read CSV: %v", rErr)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record (header only), got %d", len(records))
	}
}

func TestExportLinkNotFound(t *testing.T) {
	linkRepo := &mockLinkRepo{
		err: commonRepo.NewErrorResourceNotFound(errors.New("link not found")),
	}

	repo := &mockCorpSigningRepo{}

	s := newTestCorpSigningService(repo, linkRepo)
	_, err := s.Export("user1", "link1", false, "")
	if err == nil {
		t.Fatal("Export should return error when link not found")
	}
}

func TestExportNoPermission(t *testing.T) {
	linkRepo := &mockLinkRepo{
		link: domain.Link{Id: "link1", Submitter: "other_user"},
	}

	repo := &mockCorpSigningRepo{}

	s := newTestCorpSigningService(repo, linkRepo)
	_, err := s.Export("user1", "link1", false, "")
	if err == nil {
		t.Fatal("Export should return error when user has no permission")
	}
}

func TestExportMultiPage(t *testing.T) {
	linkRepo := &mockLinkRepo{
		link: domain.Link{Id: "link1", Submitter: "user1"},
	}

	page1Data := make([]repository.CorpSigningSummary, 200)
	for i := range page1Data {
		page1Data[i] = repository.CorpSigningSummary{
			Id:   "cs" + string(rune('A'+i)),
			Date: "2026-08-20 10:23:45",
			Link: domain.LinkInfo{
				Id:      "link1",
				CLAInfo: domain.CLAInfo{CLAId: "cla1", Language: dp.CreateLanguage("zh")},
			},
			Rep: domain.Representative{
				Name:      dp.CreateName("rep"),
				EmailAddr: dp.CreateEmailAddr("rep@test.com"),
			},
			Corp:  domain.Corporation{Name: dp.CreateCorpName("corp")},
			Admin: domain.Manager{},
		}
	}

	page2Data := make([]repository.CorpSigningSummary, 5)
	for i := range page2Data {
		page2Data[i] = page1Data[0]
	}

	repo := &mockCorpSigningRepo{
		pages: map[int]repository.CorpSigningSummaryPage{
			1: {Total: 205, Data: page1Data},
			2: {Total: 205, Data: page2Data},
		},
	}

	s := newTestCorpSigningService(repo, linkRepo)
	data, err := s.Export("user1", "link1", false, "")
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	body := bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	r := csv.NewReader(bytes.NewReader(body))
	records, rErr := r.ReadAll()
	if rErr != nil {
		t.Fatalf("failed to read CSV: %v", rErr)
	}

	if len(records) != 206 {
		t.Errorf("expected 206 records (header + 205 rows), got %d", len(records))
	}
}
