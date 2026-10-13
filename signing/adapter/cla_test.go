package adapter

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestFormatCLADate(t *testing.T) {
	tests := []struct {
		name string
		unix int64
		want string
	}{
		{"zero (历史遗留)", 0, "1970-01-01"},
		{"negative", -1, "1970-01-01"},
		{"valid timestamp", 1726099200, "2024-09-12"},
		{"valid timestamp 2", 1704067200, "2024-01-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCLADate(tt.unix)
			if got != tt.want {
				t.Errorf("formatCLADate(%d) = %s, want %s", tt.unix, got, tt.want)
			}
		})
	}
}

func TestToCLADetailUpdatedAt(t *testing.T) {
	adapter := &claAdatper{}

	ts := time.Now().Unix()
	dto := []app.CLADTO{
		{Id: "0", URL: "http://example.com/a.pdf", Language: "zh", UpdatedAt: ts},
		{Id: "1", URL: "http://example.com/b.pdf", Language: "en", UpdatedAt: 0},
	}

	result := adapter.toCLADetail(dto)

	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}

	// first item: valid timestamp
	wantFirst := formatCLADate(ts)
	if result[0].UpdatedAt != wantFirst {
		t.Errorf("result[0].UpdatedAt = %s, want %s", result[0].UpdatedAt, wantFirst)
	}
	if result[0].CLAId != "0" {
		t.Errorf("result[0].CLAId = %s, want 0", result[0].CLAId)
	}

	// second item: zero timestamp -> 1970-01-01
	if result[1].UpdatedAt != "1970-01-01" {
		t.Errorf("result[1].UpdatedAt = %s, want 1970-01-01", result[1].UpdatedAt)
	}
}

func TestToCLADetailEmpty(t *testing.T) {
	adapter := &claAdatper{}
	result := adapter.toCLADetail([]app.CLADTO{})

	if len(result) != 0 {
		t.Errorf("expected 0 items, got %d", len(result))
	}
}

func TestToCLAFields(t *testing.T) {
	fields := []domain.Field{
		{
			Id:       "1",
			Required: true,
			CLAField: dp.CLAField{
				Type:  "name",
				Desc:  "description1",
				Title: "Name",
			},
		},
		{
			Id:       "2",
			Required: false,
			CLAField: dp.CLAField{
				Type:  "email",
				Desc:  "description2",
				Title: "Email",
			},
		},
	}

	result := toCLAFields(fields)

	if len(result) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(result))
	}

	if result[0].ID != "1" || result[0].Title != "Name" || result[0].Type != "name" {
		t.Errorf("result[0] = %+v", result[0])
	}
	if !result[0].Required {
		t.Error("result[0] should be required")
	}

	if result[1].ID != "2" || result[1].Title != "Email" || result[1].Type != "email" {
		t.Errorf("result[1] = %+v", result[1])
	}
	if result[1].Required {
		t.Error("result[1] should not be required")
	}
}

func TestToCLAFieldsEmpty(t *testing.T) {
	result := toCLAFields([]domain.Field{})

	if len(result) != 0 {
		t.Errorf("expected 0 fields, got %d", len(result))
	}
}

func TestDiffPreviewNotAllowedSource(t *testing.T) {
	adapter := &claAdatper{
		claPDFSource: []string{"https://gitee.com"},
	}

	opt := &models.CLADiffPreviewOpt{
		URL:      "https://evil.com/x.pdf",
		Type:     "corporation",
		Language: "zh",
	}

	_, merr := adapter.DiffPreview("user1", "link1", opt)
	if merr == nil {
		t.Fatal("expected error for not allowed source")
	}
	if !merr.IsErrorOf(models.ErrNotAllowedCLAPDFSource) {
		t.Errorf("expected ErrNotAllowedCLAPDFSource, got %s", merr.ErrCode())
	}
}

func TestDiffPreviewInvalidType(t *testing.T) {
	adapter := &claAdatper{
		claPDFSource:         []string{"https://gitee.com"},
		fileTypeOfCLAContent: "pdf",
		maxSizeOfCLAContent:  2 << 20,
	}

	opt := &models.CLADiffPreviewOpt{
		URL:      "https://gitee.com/x.pdf",
		Type:     "invalid_type",
		Language: "zh",
	}

	_, merr := adapter.DiffPreview("user1", "link1", opt)
	if merr == nil {
		t.Fatal("expected error for invalid type")
	}
	if !merr.IsErrorOf(models.ErrBadRequestParameter) {
		t.Errorf("expected ErrBadRequestParameter, got %s", merr.ErrCode())
	}
}

func TestDiffPreviewInvalidLanguage(t *testing.T) {
	adapter := &claAdatper{
		claPDFSource:         []string{"https://gitee.com"},
		fileTypeOfCLAContent: "pdf",
		maxSizeOfCLAContent:  2 << 20,
	}

	opt := &models.CLADiffPreviewOpt{
		URL:      "https://gitee.com/x.pdf",
		Type:     "corporation",
		Language: "fr",
	}

	_, merr := adapter.DiffPreview("user1", "link1", opt)
	if merr == nil {
		t.Fatal("expected error for invalid language")
	}
	if !merr.IsErrorOf(models.ErrBadRequestParameter) {
		t.Errorf("expected ErrBadRequestParameter, got %s", merr.ErrCode())
	}
}

func TestDiffPreviewIsNew(t *testing.T) {
	srv := newTestPDFServer()
	defer srv.Close()

	adapter := &claAdatper{
		s: &mockCLAService{
			diffPreviewId:      "",
			diffPreviewOldFile: "",
			diffPreviewErr:     nil,
		},
		claPDFSource:         []string{srv.URL},
		fileTypeOfCLAContent: "pdf",
		maxSizeOfCLAContent:  2 << 20,
		pythonBin:            "/bin/true",
		pythonRetryTimes:     3,
	}

	opt := &models.CLADiffPreviewOpt{
		URL:      srv.URL + "/test.pdf",
		Type:     "corporation",
		Language: "Chinese",
	}

	result, merr := adapter.DiffPreview("user1", "link1", opt)
	if merr != nil {
		t.Fatalf("unexpected error: %v", merr)
	}
	if !result.IsNew {
		t.Error("expected IsNew=true")
	}
	if result.DiffHTML != "" {
		t.Error("expected empty DiffHTML for new CLA")
	}
}

func TestDiffPreviewGenDiffFailed(t *testing.T) {
	srv := newTestPDFServer()
	defer srv.Close()

	oldFile, err := os.CreateTemp("", "old_cla_*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	oldFile.Close()
	defer os.Remove(oldFile.Name())

	adapter := &claAdatper{
		s: &mockCLAService{
			diffPreviewId:      "cla1",
			diffPreviewOldFile: oldFile.Name(),
			diffPreviewErr:     nil,
		},
		claPDFSource:         []string{srv.URL},
		fileTypeOfCLAContent: "pdf",
		maxSizeOfCLAContent:  2 << 20,
		pythonBin:            "/nonexistent/python-bin",
		pythonRetryTimes:     3,
	}

	opt := &models.CLADiffPreviewOpt{
		URL:      srv.URL + "/test.pdf",
		Type:     "corporation",
		Language: "Chinese",
	}

	_, merr := adapter.DiffPreview("user1", "link1", opt)
	if merr == nil {
		t.Fatal("expected error for diff generation failure")
	}
	if !merr.IsErrorOf(models.ErrGenDiffFailed) {
		t.Errorf("expected ErrGenDiffFailed, got %s", merr.ErrCode())
	}
}

func TestDiffPreviewSuccess(t *testing.T) {
	srv := newTestPDFServer()
	defer srv.Close()

	oldFile, err := os.CreateTemp("", "old_cla_*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	oldFile.Close()
	defer os.Remove(oldFile.Name())

	adapter := &claAdatper{
		s: &mockCLAService{
			diffPreviewId:      "cla1",
			diffPreviewOldFile: oldFile.Name(),
			diffPreviewErr:     nil,
		},
		claPDFSource:         []string{srv.URL},
		fileTypeOfCLAContent: "pdf",
		maxSizeOfCLAContent:  2 << 20,
		pythonBin:            "/bin/true",
		pythonRetryTimes:     3,
	}

	opt := &models.CLADiffPreviewOpt{
		URL:      srv.URL + "/test.pdf",
		Type:     "corporation",
		Language: "Chinese",
	}

	result, merr := adapter.DiffPreview("user1", "link1", opt)
	if merr != nil {
		t.Fatalf("unexpected error: %v", merr)
	}
	if result.IsNew {
		t.Error("expected IsNew=false")
	}
}

func TestRunGenerateDiffFailure(t *testing.T) {
	adapter := &claAdatper{
		pythonBin:        "/nonexistent/python-bin",
		pythonRetryTimes: 3,
	}

	err := adapter.runGenerateDiff("old.pdf", "new.pdf", "out.html")
	if err == nil {
		t.Fatal("expected error for non-existent python")
	}
}

func TestRunGenerateDiffSuccess(t *testing.T) {
	adapter := &claAdatper{
		pythonBin:        "/bin/true",
		pythonRetryTimes: 3,
	}

	err := adapter.runGenerateDiff("old.pdf", "new.pdf", "out.html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunGenerateDiffDefaultRetry(t *testing.T) {
	adapter := &claAdatper{
		pythonBin:        "/bin/true",
		pythonRetryTimes: 0,
	}

	err := adapter.runGenerateDiff("old.pdf", "new.pdf", "out.html")
	if err != nil {
		t.Fatalf("unexpected error with default retry: %v", err)
	}
}

func TestCorpCLATemplatePDFServiceError(t *testing.T) {
	adapter := &claAdatper{
		s: &mockCLAService{
			corpDetailErr: domain.NewDomainError(domain.ErrorCodeCLANotCorp),
		},
	}

	_, merr := adapter.CorpCLATemplatePDF("user1", "link1", "cla1")
	if merr == nil {
		t.Fatal("expected error from service")
	}
	if !merr.IsErrorOf(models.ErrCLANotCorp) {
		t.Errorf("expected ErrCLANotCorp, got %s", merr.ErrCode())
	}
}

func TestCorpCLATemplatePDFGenNil(t *testing.T) {
	adapter := &claAdatper{
		s: &mockCLAService{
			corpDetail: app.CLADetailDTO{
				Id:        "cla1",
				Language:  "English",
				LocalFile: "/tmp/test.pdf",
			},
		},
	}

	_, merr := adapter.CorpCLATemplatePDF("user1", "link1", "cla1")
	if merr == nil {
		t.Fatal("expected error for nil pdf generator")
	}
	if !merr.IsErrorOf(models.ErrGenTemplatePDFFailed) {
		t.Errorf("expected ErrGenTemplatePDFFailed, got %s", merr.ErrCode())
	}
}

func TestCorpCLATemplatePDFGenError(t *testing.T) {
	tmpDir := t.TempDir()
	pdf.InitPDFGenerator(&pdf.Config{
		PDFOutDir: tmpDir,
		PythonBin: "/bin/true",
	})

	adapter := &claAdatper{
		s: &mockCLAService{
			corpDetail: app.CLADetailDTO{
				Id:        "cla1",
				Language:  "french",
				LocalFile: "/tmp/test.pdf",
			},
		},
	}

	_, merr := adapter.CorpCLATemplatePDF("user1", "link1", "cla1")
	if merr == nil {
		t.Fatal("expected error for unknown language")
	}
	if !merr.IsErrorOf(models.ErrGenTemplatePDFFailed) {
		t.Errorf("expected ErrGenTemplatePDFFailed, got %s", merr.ErrCode())
	}
}

func TestCorpCLATemplatePDFSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	pdf.InitPDFGenerator(&pdf.Config{
		PDFOutDir: tmpDir,
		PythonBin: "/bin/true",
	})

	adapter := &claAdatper{
		s: &mockCLAService{
			corpDetail: app.CLADetailDTO{
				Id:        "cla1",
				Language:  "English",
				LocalFile: "/tmp/test.pdf",
			},
		},
	}

	outfile, merr := adapter.CorpCLATemplatePDF("user1", "link1", "cla1")
	if merr != nil {
		t.Fatalf("unexpected error: %v", merr)
	}
	if outfile == "" {
		t.Error("expected non-empty outfile path")
	}
}

func newTestPDFServer() *httptest.Server {
	body := []byte("dummy pdf content")
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
}
