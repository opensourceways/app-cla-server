package domain

import "testing"

func TestConfigSetDefault(t *testing.T) {
	cfg := Config{}

	cfg.SetDefault()

	if cfg.VerificationCodeExpiry != 300 {
		t.Errorf("VerificationCodeExpiry: got %d, want 300", cfg.VerificationCodeExpiry)
	}
	if cfg.MaxNumOfEmployeeManager != 5 {
		t.Errorf("MaxNumOfEmployeeManager: got %d, want 5", cfg.MaxNumOfEmployeeManager)
	}
	if len(cfg.SourceOfCLAPDF) != 3 {
		t.Errorf("SourceOfCLAPDF length: got %d, want 3", len(cfg.SourceOfCLAPDF))
	}
	if cfg.MaxSizeOfCLAContent != 2<<20 {
		t.Errorf("MaxSizeOfCLAContent: got %d, want %d", cfg.MaxSizeOfCLAContent, 2<<20)
	}
	if cfg.FileTypeOfCLAContent != "pdf" {
		t.Errorf("FileTypeOfCLAContent: got %s, want pdf", cfg.FileTypeOfCLAContent)
	}
	if cfg.AccessTokenExpiry != 3600 {
		t.Errorf("AccessTokenExpiry: got %d, want 3600", cfg.AccessTokenExpiry)
	}
	if cfg.MaxNumOfFailedLogin != 5 {
		t.Errorf("MaxNumOfFailedLogin: got %d, want 5", cfg.MaxNumOfFailedLogin)
	}
	if cfg.IntervalOfCreatingVC != 60 {
		t.Errorf("IntervalOfCreatingVC: got %d, want 60", cfg.IntervalOfCreatingVC)
	}
	if cfg.CommunityManagerLinkId != "fake_link" {
		t.Errorf("CommunityManagerLinkId: got %s, want fake_link", cfg.CommunityManagerLinkId)
	}
	if cfg.DefaultGracePeriodDays != 30 {
		t.Errorf("DefaultGracePeriodDays: got %d, want 30", cfg.DefaultGracePeriodDays)
	}
}

func TestConfigSetDefaultPreservesExisting(t *testing.T) {
	cfg := Config{
		VerificationCodeExpiry:  600,
		MaxNumOfEmployeeManager: 10,
		SourceOfCLAPDF:          []string{"https://custom.com"},
		MaxSizeOfCLAContent:     1024,
		FileTypeOfCLAContent:    "html",
		AccessTokenExpiry:       7200,
		MaxNumOfFailedLogin:     3,
		IntervalOfCreatingVC:    120,
		CommunityManagerLinkId:  "my_link",
		DefaultGracePeriodDays:  45,
	}

	cfg.SetDefault()

	if cfg.VerificationCodeExpiry != 600 {
		t.Errorf("VerificationCodeExpiry should be preserved: got %d, want 600", cfg.VerificationCodeExpiry)
	}
	if cfg.MaxNumOfEmployeeManager != 10 {
		t.Errorf("MaxNumOfEmployeeManager should be preserved: got %d, want 10", cfg.MaxNumOfEmployeeManager)
	}
	if len(cfg.SourceOfCLAPDF) != 1 {
		t.Errorf("SourceOfCLAPDF should be preserved: got %d, want 1", len(cfg.SourceOfCLAPDF))
	}
	if cfg.MaxSizeOfCLAContent != 1024 {
		t.Errorf("MaxSizeOfCLAContent should be preserved: got %d, want 1024", cfg.MaxSizeOfCLAContent)
	}
	if cfg.FileTypeOfCLAContent != "html" {
		t.Errorf("FileTypeOfCLAContent should be preserved: got %s, want html", cfg.FileTypeOfCLAContent)
	}
	if cfg.AccessTokenExpiry != 7200 {
		t.Errorf("AccessTokenExpiry should be preserved: got %d, want 7200", cfg.AccessTokenExpiry)
	}
	if cfg.MaxNumOfFailedLogin != 3 {
		t.Errorf("MaxNumOfFailedLogin should be preserved: got %d, want 3", cfg.MaxNumOfFailedLogin)
	}
	if cfg.IntervalOfCreatingVC != 120 {
		t.Errorf("IntervalOfCreatingVC should be preserved: got %d, want 120", cfg.IntervalOfCreatingVC)
	}
	if cfg.CommunityManagerLinkId != "my_link" {
		t.Errorf("CommunityManagerLinkId should be preserved: got %s, want my_link", cfg.CommunityManagerLinkId)
	}
	if cfg.DefaultGracePeriodDays != 45 {
		t.Errorf("DefaultGracePeriodDays should be preserved: got %d, want 45", cfg.DefaultGracePeriodDays)
	}
}

func TestConfigSetDefaultZeroValues(t *testing.T) {
	cfg := Config{
		DefaultGracePeriodDays: 0,
	}

	cfg.SetDefault()

	if cfg.DefaultGracePeriodDays != 30 {
		t.Errorf("DefaultGracePeriodDays 0 should default to 30: got %d", cfg.DefaultGracePeriodDays)
	}
}

func TestConfigInvalidCorpEmailDomains(t *testing.T) {
	cfg := Config{InvalidCorpEmailDomain: "a,b,c"}
	domains := cfg.InvalidCorpEmailDomains()
	if len(domains) != 3 {
		t.Errorf("InvalidCorpEmailDomains length: got %d, want 3", len(domains))
	}
	if domains[0] != "a" || domains[1] != "b" || domains[2] != "c" {
		t.Errorf("InvalidCorpEmailDomains content: got %v", domains)
	}
}

func TestConfigGetIntervalOfCreatingVC(t *testing.T) {
	cfg := Config{IntervalOfCreatingVC: 60}
	d := cfg.GetIntervalOfCreatingVC()
	if d.Seconds() != 60 {
		t.Errorf("GetIntervalOfCreatingVC: got %f, want 60", d.Seconds())
	}
}
