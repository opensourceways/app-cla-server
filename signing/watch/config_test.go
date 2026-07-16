package watch

import (
	"testing"
	"time"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

func TestGetEffectiveGracePeriodDays(t *testing.T) {
	impl := &notifyAdminWatchImpl{defaultGracePeriodDays: 30}
	intPtr := func(v int) *int { return &v }

	tests := []struct {
		name       string
		linkDays   *int
		defaultDay int
		want       int
	}{
		{"link nil (字段缺失) uses default", nil, 30, 30},
		{"link override positive", intPtr(15), 30, 15},
		{"link zero returns zero", intPtr(0), 30, 0},
		{"link negative uses default", intPtr(-1), 30, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := &repository.LinkCLA{GracePeriodDays: tt.linkDays}
			impl.defaultGracePeriodDays = tt.defaultDay
			got := impl.getEffectiveGracePeriodDays(link)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIsCorpSigningLatest(t *testing.T) {
	impl := &notifyAdminWatchImpl{}

	clas := []domain.CLA{
		{Id: "10", Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")},
		{Id: "11", Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("en")},
	}

	signedInfo := domain.CLAInfo{CLAId: "10", Language: dp.CreateLanguage("en")}
	if !impl.isCorpSigningLatest(clas, signedInfo) {
		t.Error("isCorpSigningLatest should return true for matching CLA")
	}

	signedInfo2 := domain.CLAInfo{CLAId: "9", Language: dp.CreateLanguage("en")}
	if impl.isCorpSigningLatest(clas, signedInfo2) {
		t.Error("isCorpSigningLatest should return false for outdated CLA")
	}
}

func TestGetLatestCorpClaId(t *testing.T) {
	impl := &notifyAdminWatchImpl{}

	link := &repository.LinkCLA{
		Clas: []domain.CLA{
			{Id: "10", Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")},
			{Id: "11", Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("en")},
			{Id: "12", Type: dp.CLATypeCorp, Language: dp.CreateLanguage("zh")},
		},
	}

	if id := impl.getLatestCorpClaId(link, dp.CreateLanguage("en")); id != "10" {
		t.Errorf("getLatestCorpClaId en: got %s, want 10", id)
	}
	if id := impl.getLatestCorpClaId(link, dp.CreateLanguage("zh")); id != "12" {
		t.Errorf("getLatestCorpClaId zh: got %s, want 12", id)
	}
	if id := impl.getLatestCorpClaId(link, dp.CreateLanguage("fr")); id != "" {
		t.Errorf("getLatestCorpClaId fr: got %s, want empty", id)
	}
}

func TestGetLatestIndividualClaId(t *testing.T) {
	impl := &notifyAdminWatchImpl{}

	link := &repository.LinkCLA{
		Clas: []domain.CLA{
			{Id: "10", Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")},
			{Id: "11", Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("en")},
			{Id: "12", Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("zh")},
		},
	}

	if id := impl.getLatestIndividualClaId(link, dp.CreateLanguage("en")); id != "11" {
		t.Errorf("getLatestIndividualClaId en: got %s, want 11", id)
	}
	if id := impl.getLatestIndividualClaId(link, dp.CreateLanguage("zh")); id != "12" {
		t.Errorf("getLatestIndividualClaId zh: got %s, want 12", id)
	}
	if id := impl.getLatestIndividualClaId(link, dp.CreateLanguage("fr")); id != "" {
		t.Errorf("getLatestIndividualClaId fr: got %s, want empty", id)
	}
}

func TestNotifyAdminConfigSetDefault(t *testing.T) {
	cfg := NotifyAdminConfig{}
	cfg.SetDefault()

	if cfg.SendEmailInterval != 10 {
		t.Errorf("SendEmailInterval: got %d, want 10", cfg.SendEmailInterval)
	}
	if cfg.NotifyCorpAdminInterval != 86400 {
		t.Errorf("NotifyCorpAdminInterval: got %d, want 86400", cfg.NotifyCorpAdminInterval)
	}
}

func TestNotifyAdminConfigSetDefaultPreserves(t *testing.T) {
	cfg := NotifyAdminConfig{SendEmailInterval: 20, NotifyCorpAdminInterval: 600}
	cfg.SetDefault()

	if cfg.SendEmailInterval != 20 {
		t.Errorf("SendEmailInterval should be preserved: got %d, want 20", cfg.SendEmailInterval)
	}
	if cfg.NotifyCorpAdminInterval != 600 {
		t.Errorf("NotifyCorpAdminInterval should be preserved: got %d, want 600", cfg.NotifyCorpAdminInterval)
	}
}

func TestNotifyAdminConfigGenSendEmailInterval(t *testing.T) {
	cfg := NotifyAdminConfig{SendEmailInterval: 5}
	d := cfg.genSendEmailInterval()
	if d != 5*time.Second {
		t.Errorf("genSendEmailInterval: got %v, want 5s", d)
	}
}

func TestNotifyAdminConfigGenNotifyCorpAdminInterval(t *testing.T) {
	cfg := NotifyAdminConfig{NotifyCorpAdminInterval: 100}
	d := cfg.genNotifyCorpAdminInterval()
	if d != 100*time.Second {
		t.Errorf("genNotifyCorpAdminInterval: got %v, want 100s", d)
	}
}

func TestCLAUpdateConfigSetDefault(t *testing.T) {
	cfg := CLAUpdateConfig{}
	cfg.SetDefault()

	if cfg.MsgChannelSize != 100 {
		t.Errorf("MsgChannelSize: got %d, want 100", cfg.MsgChannelSize)
	}
	if cfg.PythonRetryTimes != 3 {
		t.Errorf("PythonRetryTimes: got %d, want 3", cfg.PythonRetryTimes)
	}
	if cfg.GenAllDiffInterval != 600 {
		t.Errorf("GenAllDiffInterval: got %d, want 600", cfg.GenAllDiffInterval)
	}
}

func TestCLAUpdateConfigSetDefaultPreserves(t *testing.T) {
	cfg := CLAUpdateConfig{MsgChannelSize: 200, PythonRetryTimes: 5, GenAllDiffInterval: 300}
	cfg.SetDefault()

	if cfg.MsgChannelSize != 200 {
		t.Errorf("MsgChannelSize preserved: got %d, want 200", cfg.MsgChannelSize)
	}
	if cfg.PythonRetryTimes != 5 {
		t.Errorf("PythonRetryTimes preserved: got %d, want 5", cfg.PythonRetryTimes)
	}
	if cfg.GenAllDiffInterval != 300 {
		t.Errorf("GenAllDiffInterval preserved: got %d, want 300", cfg.GenAllDiffInterval)
	}
}

func TestCLAUpdateConfigGenAllDiffInterval(t *testing.T) {
	cfg := CLAUpdateConfig{GenAllDiffInterval: 120}
	d := cfg.genAllDiffInterval()
	if d != 120*time.Second {
		t.Errorf("genAllDiffInterval: got %v, want 120s", d)
	}
}

func TestWatchConfigSetDefault(t *testing.T) {
	cfg := Config{}
	cfg.SetDefault()

	if cfg.Interval != 3600 {
		t.Errorf("Interval: got %d, want 3600", cfg.Interval)
	}
}

func TestWatchConfigSetDefaultPreserves(t *testing.T) {
	cfg := Config{Interval: 7200}
	cfg.SetDefault()

	if cfg.Interval != 7200 {
		t.Errorf("Interval preserved: got %d, want 7200", cfg.Interval)
	}
}

func TestWatchConfigConfigItems(t *testing.T) {
	cfg := Config{}
	items := cfg.ConfigItems()
	if len(items) != 2 {
		t.Errorf("ConfigItems length: got %d, want 2", len(items))
	}
}

func TestWatchConfigIntervalDuration(t *testing.T) {
	cfg := Config{Interval: 60}
	d := cfg.intervalDuration()
	if d != 60*time.Second {
		t.Errorf("intervalDuration: got %v, want 60s", d)
	}
}
