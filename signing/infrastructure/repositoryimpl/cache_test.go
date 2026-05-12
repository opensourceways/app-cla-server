package repositoryimpl

import (
	"testing"
)

func TestPageKey(t *testing.T) {
	key := pageKey("link-1", 1, 20, false, "")
	expected := corpSigningPageCachePrefix + "link-1:1:20:false:"
	if key != expected {
		t.Errorf("expected '%s', got '%s'", expected, key)
	}

	key2 := pageKey("link-2", 2, 10, true, "search-term")
	expected2 := corpSigningPageCachePrefix + "link-2:2:10:true:search-term"
	if key2 != expected2 {
		t.Errorf("expected '%s', got '%s'", expected2, key2)
	}
}

func TestLinkPattern(t *testing.T) {
	pattern := linkPattern("link-1")
	expected := corpSigningPageCachePrefix + "link-1:*"
	if pattern != expected {
		t.Errorf("expected '%s', got '%s'", expected, pattern)
	}
}

func TestCorpSigningPageCacheConstants(t *testing.T) {
	if corpSigningPageCachePrefix != "corp_signing:page:" {
		t.Errorf("unexpected prefix: '%s'", corpSigningPageCachePrefix)
	}
	if corpSigningPageCacheTTL == 0 {
		t.Error("expected non-zero TTL")
	}
}

func TestCachedCorpSigningSummary(t *testing.T) {
	s := cachedCorpSigningSummary{
		Id:       "cs-1",
		Date:     "2024-01-01",
		HasPDF:   true,
		LinkId:   "link-1",
		CLAId:    "cla-1",
		Language: "en",
		RepName:  "CEO",
		RepEmail: "ceo@t.com",
	}
	if s.Id != "cs-1" || s.HasPDF != true {
		t.Errorf("cachedCorpSigningSummary mismatch")
	}
}

func TestCachedCorpSigningPage(t *testing.T) {
	page := cachedCorpSigningPage{
		Total: 10,
		Data:  []cachedCorpSigningSummary{{Id: "cs-1"}},
	}
	if page.Total != 10 || len(page.Data) != 1 {
		t.Errorf("cachedCorpSigningPage mismatch")
	}
}
