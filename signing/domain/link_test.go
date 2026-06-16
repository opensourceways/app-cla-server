package domain

import "testing"

func TestGetEffectiveGracePeriodDays(t *testing.T) {
	tests := []struct {
		name        string
		linkDays    int
		defaultDays int
		want        int
	}{
		{"negative means use default", -1, 30, 30},
		{"negative with zero default", -1, 0, 0},
		{"zero means zero (never block)", 0, 30, 0},
		{"positive overrides default", 15, 30, 15},
		{"positive overrides zero default", 10, 0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := &Link{GracePeriodDays: tt.linkDays}
			got := link.GetEffectiveGracePeriodDays(tt.defaultDays)
			if got != tt.want {
				t.Errorf("GetEffectiveGracePeriodDays(%d) with linkDays=%d = %d, want %d",
					tt.defaultDays, tt.linkDays, got, tt.want)
			}
		})
	}
}

func TestCorpSigningSetLatestClaId(t *testing.T) {
	cs := &CorpSigning{
		Link:           LinkInfo{CLAInfo: CLAInfo{CLAId: "3"}},
		PendingCLAId:   "5",
		ClaNotifyCount: 2,
		ClaNotifyTime:  1000,
		Logs:           []CorpSigningLog{},
	}

	err := cs.SetLatestClaId("5")
	if err != nil {
		t.Fatalf("SetLatestClaId returned unexpected error: %v", err)
	}

	if cs.Link.CLAId != "5" {
		t.Errorf("CLAId not updated: got %s, want 5", cs.Link.CLAId)
	}
	if cs.PendingCLAId != "" {
		t.Errorf("PendingCLAId not cleared: got %s, want empty", cs.PendingCLAId)
	}
	if cs.ClaNotifyCount != 0 {
		t.Errorf("ClaNotifyCount not reset: got %d, want 0", cs.ClaNotifyCount)
	}
	if cs.ClaNotifyTime != 0 {
		t.Errorf("ClaNotifyTime not reset: got %d, want 0", cs.ClaNotifyTime)
	}
	if len(cs.Logs) != 1 {
		t.Fatalf("Logs length: got %d, want 1", len(cs.Logs))
	}
	if cs.Logs[0].Action != "agree" {
		t.Errorf("Log action: got %s, want agree", cs.Logs[0].Action)
	}
	if cs.Logs[0].CLAId != "5" {
		t.Errorf("Log CLAId: got %s, want 5", cs.Logs[0].CLAId)
	}
}

func TestCorpSigningSetLatestClaIdAlreadyLatest(t *testing.T) {
	cs := &CorpSigning{
		Link: LinkInfo{CLAInfo: CLAInfo{CLAId: "5"}},
	}

	err := cs.SetLatestClaId("5")
	if err == nil {
		t.Error("SetLatestClaId should return error when already latest")
	}
}

func TestIndividualSigningAgreeNewCLA(t *testing.T) {
	is := &IndividualSigning{
		Link:           LinkInfo{CLAInfo: CLAInfo{CLAId: "3"}},
		Logs:           []IndividualSigningLog{{Date: "2025-01-01", ClaId: "3", Action: "sign"}},
		ClaNotify:      "5",
		ClaNotifyCount: 3,
		ClaNotifyTime:  2000,
	}

	err := is.AgreeNewCLA("5")
	if err != nil {
		t.Fatalf("AgreeNewCLA returned unexpected error: %v", err)
	}

	if is.Link.CLAId != "5" {
		t.Errorf("CLAId not updated: got %s, want 5", is.Link.CLAId)
	}
	if is.ClaNotify != "" {
		t.Errorf("ClaNotify not reset: got %s, want empty", is.ClaNotify)
	}
	if is.ClaNotifyCount != 0 {
		t.Errorf("ClaNotifyCount not reset: got %d, want 0", is.ClaNotifyCount)
	}
	if is.ClaNotifyTime != 0 {
		t.Errorf("ClaNotifyTime not reset: got %d, want 0", is.ClaNotifyTime)
	}
	lastLog := is.Logs[len(is.Logs)-1]
	if lastLog.Action != "agree" {
		t.Errorf("Last log action: got %s, want agree", lastLog.Action)
	}
	if lastLog.ClaId != "5" {
		t.Errorf("Last log ClaId: got %s, want 5", lastLog.ClaId)
	}
}

func TestIndividualSigningAgreeNewCLAAlreadyLatest(t *testing.T) {
	is := &IndividualSigning{
		Link: LinkInfo{CLAInfo: CLAInfo{CLAId: "5"}},
	}

	err := is.AgreeNewCLA("5")
	if err == nil {
		t.Error("AgreeNewCLA should return error when already latest")
	}
}
