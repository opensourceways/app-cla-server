package claservice

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func newTestCLA(id, claType string, langStr string) domain.CLA {
	return domain.CLA{
		Id:       id,
		Type:     dp.CreateCLAType(claType),
		Language: dp.CreateLanguage(langStr),
	}
}

func TestLinkCacheContains(t *testing.T) {
	lc := &linkCache{
		cache: map[string][]domain.CLA{
			"link1": {newTestCLA("cla1", "corporation", "en")},
		},
	}

	if !lc.contains("link1", "cla1") {
		t.Error("expected true for existing cla")
	}
	if lc.contains("link1", "nonexistent") {
		t.Error("expected false for non-existent cla id")
	}
	if lc.contains("nonexistent", "cla1") {
		t.Error("expected false for non-existent link id")
	}
}

func TestLinkCacheGetClaId(t *testing.T) {
	lc := &linkCache{
		cache: map[string][]domain.CLA{
			"link1": {
				newTestCLA("id0", "corporation", "en"),
				newTestCLA("id1", "individual", "en"),
			},
		},
	}

	id := lc.getClaId("link1", dp.CLATypeCorp, dp.CreateLanguage("en"))
	if id != "id0" {
		t.Errorf("expected 'id0', got '%s'", id)
	}

	id = lc.getClaId("link1", dp.CLATypeIndividual, dp.CreateLanguage("en"))
	if id != "id1" {
		t.Errorf("expected 'id1', got '%s'", id)
	}

	id = lc.getClaId("link1", dp.CreateCLAType("unknown"), dp.CreateLanguage("en"))
	if id != "" {
		t.Errorf("expected '', got '%s'", id)
	}

	id = lc.getClaId("nonexistent", dp.CLATypeCorp, dp.CreateLanguage("en"))
	if id != "" {
		t.Errorf("expected '', got '%s'", id)
	}
}

func TestLinkCacheUpdateNewLink(t *testing.T) {
	lc := &linkCache{
		cache: make(map[string][]domain.CLA),
	}

	cla := newTestCLA("new", "corporation", "en")
	lc.update("link1", &cla)

	if !lc.contains("link1", "new") {
		t.Error("expected cached CLA after update")
	}
}

func TestLinkCacheUpdateExistingType(t *testing.T) {
	lc := &linkCache{
		cache: map[string][]domain.CLA{
			"link1": {newTestCLA("old", "corporation", "en")},
		},
	}

	updated := newTestCLA("new", "corporation", "en")
	lc.update("link1", &updated)

	if lc.contains("link1", "old") {
		t.Error("expected old CLA to be replaced")
	}
	if !lc.contains("link1", "new") {
		t.Error("expected new CLA to be cached")
	}
}

func TestLinkCacheUpdateNewLanguage(t *testing.T) {
	lc := &linkCache{
		cache: map[string][]domain.CLA{
			"link1": {newTestCLA("0", "corporation", "en")},
		},
	}

	// Add different language
	newCLA := newTestCLA("1", "corporation", "zh")
	lc.update("link1", &newCLA)

	if !lc.contains("link1", "0") {
		t.Error("expected original CLA to remain")
	}
	if !lc.contains("link1", "1") {
		t.Error("expected new CLA to be added")
	}
}

func TestLinkCacheRemoveLink(t *testing.T) {
	lc := &linkCache{
		cache: map[string][]domain.CLA{
			"link1": {newTestCLA("0", "corporation", "en")},
			"link2": {newTestCLA("1", "individual", "en")},
		},
	}

	lc.removeLink("link1")
	if lc.contains("link1", "0") {
		t.Error("expected link1 to be removed")
	}
	if !lc.contains("link2", "1") {
		t.Error("expected link2 to remain")
	}
}

func TestLinkCacheRemoveCLA(t *testing.T) {
	lc := &linkCache{
		cache: map[string][]domain.CLA{
			"link1": {
				newTestCLA("0", "corporation", "en"),
				newTestCLA("1", "individual", "en"),
			},
		},
	}

	lc.removeCLA("link1", "0")
	if lc.contains("link1", "0") {
		t.Error("expected cla0 to be removed")
	}
	if !lc.contains("link1", "1") {
		t.Error("expected cla1 to remain")
	}

	lc.removeCLA("link1", "1")
	if lc.contains("link1", "1") {
		t.Error("expected cla1 to be removed")
	}
	// link1 still exists with empty slice
	if len(lc.cache["link1"]) != 0 {
		t.Errorf("expected empty slice, got %d items", len(lc.cache["link1"]))
	}
}

func TestLinkCacheRemoveCLANonexistentLink(t *testing.T) {
	lc := &linkCache{
		cache: make(map[string][]domain.CLA),
	}

	// should not panic
	lc.removeCLA("nonexistent", "0")
}
