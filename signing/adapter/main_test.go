package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func findModuleRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func TestMain(m *testing.M) {
	if root := findModuleRoot(); root != "" {
		_ = os.Chdir(root)
	}

	cfg := &dp.Config{}
	json.Unmarshal([]byte(`{
		"max_length_of_name": 100,
		"max_length_of_title": 200,
		"max_length_of_email": 200,
		"max_length_of_account": 100,
		"max_length_of_corp_name": 200,
		"cla": [
			{"type": "corporation", "fields": [{"language": "Chinese", "fields": []}, {"language": "English", "fields": []}]},
			{"type": "individual", "fields": [{"language": "Chinese", "fields": []}, {"language": "English", "fields": []}]}
		]
	}`), cfg)
	cfg.Validate()
	dp.Init(cfg)

	os.Exit(m.Run())
}

type mockCLAService struct {
	corpDetail         app.CLADetailDTO
	corpDetailErr      error
	diffPreviewId      string
	diffPreviewOldFile string
	diffPreviewErr     error
}

func (m *mockCLAService) Add(cmd *app.CmdToAddCLA) error { return nil }

func (m *mockCLAService) Update(cmd *app.CmdToUpdateCLA) error { return nil }

func (m *mockCLAService) Remove(cmd *app.CmdToRemoveCLA) error { return nil }

func (m *mockCLAService) CLALocalFilePath(idx domain.CLAIndex) string { return "" }

func (m *mockCLAService) List(userId, linkId string) ([]app.CLADTO, []app.CLADTO, error) {
	return nil, nil, nil
}

func (m *mockCLAService) CorpCLADetail(userId, linkId, claId string) (app.CLADetailDTO, error) {
	return m.corpDetail, m.corpDetailErr
}

func (m *mockCLAService) DiffPreviewCLA(userId, linkId string, claType dp.CLAType, language dp.Language) (string, string, error) {
	return m.diffPreviewId, m.diffPreviewOldFile, m.diffPreviewErr
}
