package domain

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestLinkCanDo(t *testing.T) {
	link := &Link{Submitter: "user1"}

	if err := link.CanDo("user1"); err != nil {
		t.Errorf("CanDo same user: unexpected error %v", err)
	}
	if err := link.CanDo("user2"); err == nil {
		t.Error("CanDo different user should fail")
	}
}

func TestLinkAddCLA(t *testing.T) {
	link := &Link{CLANum: 3, CLAs: []CLA{
		{Id: "1", Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")},
		{Id: "2", Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("en")},
	}}

	cla := &CLA{Type: dp.CLATypeCorp, Language: dp.CreateLanguage("zh")}
	if err := link.AddCLA(cla); err != nil {
		t.Fatalf("AddCLA new: unexpected error %v", err)
	}
	if cla.Id != "3" {
		t.Errorf("CLA.Id: got %s, want 3", cla.Id)
	}
	if link.CLANum != 4 {
		t.Errorf("CLANum: got %d, want 4", link.CLANum)
	}

	dupCla := &CLA{Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")}
	if err := link.AddCLA(dupCla); err == nil {
		t.Error("AddCLA duplicate should fail")
	}
}

func TestLinkUpdateCLA(t *testing.T) {
	link := &Link{CLANum: 3, CLAs: []CLA{
		{Id: "1", Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")},
	}}

	cla := &CLA{Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")}
	if err := link.UpdateCLA(cla); err != nil {
		t.Fatalf("UpdateCLA existing: unexpected error %v", err)
	}
	if cla.Id != "3" {
		t.Errorf("CLA.Id after update: got %s, want 3", cla.Id)
	}
	if link.CLANum != 4 {
		t.Errorf("CLANum after update: got %d, want 4", link.CLANum)
	}

	nonExistCla := &CLA{Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("zh")}
	if err := link.UpdateCLA(nonExistCla); err == nil {
		t.Error("UpdateCLA non-existing should fail")
	}
}

func TestLinkFindCLA(t *testing.T) {
	link := &Link{CLAs: []CLA{
		{Id: "1"},
		{Id: "2"},
	}}

	if link.FindCLA("1") == nil {
		t.Error("FindCLA existing should return CLA")
	}
	if link.FindCLA("3") != nil {
		t.Error("FindCLA non-existing should return nil")
	}
}

func TestLinkGetCLA(t *testing.T) {
	link := &Link{CLAs: []CLA{
		{Type: dp.CLATypeCorp, Language: dp.CreateLanguage("en")},
		{Type: dp.CLATypeIndividual, Language: dp.CreateLanguage("zh")},
	}}

	cla := link.GetCLA(dp.CLATypeCorp, dp.CreateLanguage("en"))
	if cla == nil {
		t.Error("GetCLA corp/en should find CLA")
	}

	cla2 := link.GetCLA(dp.CLATypeIndividual, dp.CreateLanguage("zh"))
	if cla2 == nil {
		t.Error("GetCLA individual/zh should find CLA")
	}

	cla3 := link.GetCLA(dp.CLATypeCorp, dp.CreateLanguage("zh"))
	if cla3 != nil {
		t.Error("GetCLA corp/zh should return nil")
	}
}

func TestLinkGetEffectiveGracePeriodDaysAllBranches(t *testing.T) {
	intPtr := func(v int) *int { return &v }

	tests := []struct {
		name        string
		linkDays    *int
		defaultDays int
		want        int
	}{
		{"nil (字段缺失) uses default", nil, 30, 30},
		{"negative uses default", intPtr(-1), 30, 30},
		{"zero returns zero", intPtr(0), 30, 0},
		{"positive overrides", intPtr(15), 30, 15},
		{"negative with zero default", intPtr(-1), 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := &Link{GracePeriodDays: tt.linkDays}
			got := link.GetEffectiveGracePeriodDays(tt.defaultDays)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
