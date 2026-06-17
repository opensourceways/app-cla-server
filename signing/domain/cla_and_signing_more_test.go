package domain

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestCLAIsMe(t *testing.T) {
	cla1 := &CLA{Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")}
	cla2 := &CLA{Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")}
	cla3 := &CLA{Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("en")}
	cla4 := &CLA{Type: dp.CLATypeCorp, Language: dp.CreateLanguage("zh")}

	if !cla1.isMe(cla2) {
		t.Error("isMe should return true for same type and language")
	}
	if cla1.isMe(cla3) {
		t.Error("isMe should return false for different type")
	}
	if cla1.isMe(cla4) {
		t.Error("isMe should return false for different language")
	}
}

func TestIndividualSigningHasSignedCLA(t *testing.T) {
	is := &IndividualSigning{Link: LinkInfo{CLAInfo: CLAInfo{CLAId: "10"}}}

	if !is.HasSignedCLA("10") {
		t.Error("HasSignedCLA should return true for matching CLAId")
	}
	if is.HasSignedCLA("11") {
		t.Error("HasSignedCLA should return false for different CLAId")
	}
}

func TestIndividualSigningAgreeNewCLANoLogs(t *testing.T) {
	is := &IndividualSigning{
		Link:           LinkInfo{CLAInfo: CLAInfo{CLAId: "3"}},
		Logs:           []IndividualSigningLog{},
		ClaNotify:      "5",
		ClaNotifyCount: 2,
		ClaNotifyTime:  1000,
	}

	err := is.AgreeNewCLA("5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if is.Link.CLAId != "5" {
		t.Errorf("CLAId: got %s, want 5", is.Link.CLAId)
	}
	if is.ClaNotify != "" {
		t.Errorf("ClaNotify: got %s, want empty", is.ClaNotify)
	}
	if is.ClaNotifyCount != 0 {
		t.Errorf("ClaNotifyCount: got %d, want 0", is.ClaNotifyCount)
	}
	if is.ClaNotifyTime != 0 {
		t.Errorf("ClaNotifyTime: got %d, want 0", is.ClaNotifyTime)
	}

	if len(is.Logs) < 2 {
		t.Fatalf("Logs length: got %d, want at least 2", len(is.Logs))
	}
	lastLog := is.Logs[len(is.Logs)-1]
	if lastLog.Action != "agree" {
		t.Errorf("last log action: got %s, want agree", lastLog.Action)
	}
	if lastLog.ClaId != "5" {
		t.Errorf("last log ClaId: got %s, want 5", lastLog.ClaId)
	}
}

func makeManager(id, email string) Manager {
	return Manager{Id: id, Representative: Representative{EmailAddr: dp.CreateEmailAddr(email)}}
}

func TestCorpSigningAddManagers(t *testing.T) {
	cs := &CorpSigning{
		Corp: Corporation{AllEmailDomains: []string{"test.com"}},
		Managers: []Manager{
			makeManager("m1", "m1@test.com"),
		},
		Admin: makeManager("admin1", "admin@test.com"),
	}

	managers := []Manager{
		makeManager("m2", "m2@test.com"),
	}

	if err := cs.AddManagers(managers); err != nil {
		t.Fatalf("AddManagers valid: unexpected error %v", err)
	}
}

func TestCorpSigningAddManagersNotSameCorp(t *testing.T) {
	cs := &CorpSigning{
		Corp:     Corporation{AllEmailDomains: []string{"test.com"}},
		Managers: []Manager{makeManager("m1", "m1@test.com")},
	}

	managers := []Manager{
		makeManager("m2", "m2@other.com"),
	}

	if err := cs.AddManagers(managers); err == nil {
		t.Error("AddManagers with different corp should fail")
	}
}

func TestCorpSigningAddManagersDuplicate(t *testing.T) {
	cs := &CorpSigning{
		Corp: Corporation{AllEmailDomains: []string{"test.com"}},
		Managers: []Manager{
			makeManager("m1", "m1@test.com"),
		},
	}

	managers := []Manager{
		makeManager("m2", "m1@test.com"),
	}

	if err := cs.AddManagers(managers); err == nil {
		t.Error("AddManagers duplicate email should fail")
	}
}

func TestCorpSigningAddManagersAdminAsManager(t *testing.T) {
	cs := &CorpSigning{
		Corp:  Corporation{AllEmailDomains: []string{"test.com"}},
		Admin: makeManager("admin1", "admin@test.com"),
	}

	managers := []Manager{
		makeManager("m1", "admin@test.com"),
	}

	if err := cs.AddManagers(managers); err == nil {
		t.Error("AddManagers admin as manager should fail")
	}
}

func TestCorpSigningRemoveManagers(t *testing.T) {
	cs := &CorpSigning{
		Managers: []Manager{
			makeManager("m1", "m1@test.com"),
			makeManager("m2", "m2@test.com"),
			makeManager("m3", "m3@test.com"),
		},
	}

	removed, err := cs.RemoveManagers([]string{"m1", "m3"})
	if err != nil {
		t.Fatalf("RemoveManagers: unexpected error %v", err)
	}
	if len(removed) != 2 {
		t.Errorf("removed count: got %d, want 2", len(removed))
	}
	if len(cs.Managers) != 1 {
		t.Errorf("remaining managers: got %d, want 1", len(cs.Managers))
	}
	if cs.Managers[0].Id != "m2" {
		t.Errorf("remaining manager Id: got %s, want m2", cs.Managers[0].Id)
	}
}

func TestCorpSigningRemoveManagersNotFound(t *testing.T) {
	cs := &CorpSigning{
		Managers: []Manager{makeManager("m1", "m1@test.com")},
	}

	_, err := cs.RemoveManagers([]string{"unknown"})
	if err == nil {
		t.Error("RemoveManagers unknown id should fail")
	}
}

func TestCorpSigningRemoveManagersAll(t *testing.T) {
	cs := &CorpSigning{
		Managers: []Manager{
			makeManager("m1", "m1@test.com"),
			makeManager("m2", "m2@test.com"),
		},
	}

	removed, err := cs.RemoveManagers([]string{"m1", "m2"})
	if err != nil {
		t.Fatalf("RemoveManagers all: unexpected error %v", err)
	}
	if len(removed) != 2 {
		t.Errorf("removed count: got %d, want 2", len(removed))
	}
	if len(cs.Managers) != 0 {
		t.Errorf("remaining managers: got %d, want 0", len(cs.Managers))
	}
}
