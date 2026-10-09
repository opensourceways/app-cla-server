package repositoryimpl

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestCorpSigningDOAutoApproveMapping(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &domain.CorpSigning{
				Id:                   "abc",
				Date:                 "2026-01-01",
				Link:                 domain.LinkInfo{Id: "link1", CLAInfo: domain.CLAInfo{CLAId: "cla1", Language: dp.CreateLanguage("en")}},
				Rep:                  domain.Representative{Name: dp.CreateName("Rep"), EmailAddr: dp.CreateEmailAddr("rep@test.com")},
				Corp:                 domain.Corporation{Name: dp.CreateCorpName("Corp"), PrimaryEmailDomain: "test.com", AllEmailDomains: []string{"test.com"}},
				AutoApproveEmployees: tt.enabled,
			}

			do := toCorpSigningDO(cs)
			if do.AutoApproveEmployees != tt.enabled {
				t.Errorf("toCorpSigningDO: AutoApproveEmployees got %v, want %v", do.AutoApproveEmployees, tt.enabled)
			}

			back := do.toCorpSigning()
			if back.AutoApproveEmployees != tt.enabled {
				t.Errorf("toCorpSigning: AutoApproveEmployees got %v, want %v", back.AutoApproveEmployees, tt.enabled)
			}
		})
	}
}

func TestCorpSigningDOAutoApproveDefaultFalse(t *testing.T) {
	do := corpSigningDO{}
	cs := do.toCorpSigning()
	if cs.AutoApproveEmployees {
		t.Error("missing auto_approve_employees field should read as false (zero value)")
	}
}

func TestCorpSigningDOForMigrateAutoApproveMapping(t *testing.T) {
	cs := &domain.CorpSigning{
		Link:                 domain.LinkInfo{Id: "link1", CLAInfo: domain.CLAInfo{CLAId: "cla1", Language: dp.CreateLanguage("en")}},
		Rep:                  domain.Representative{Name: dp.CreateName("Rep"), EmailAddr: dp.CreateEmailAddr("rep@test.com")},
		Corp:                 domain.Corporation{Name: dp.CreateCorpName("Corp"), PrimaryEmailDomain: "test.com", AllEmailDomains: []string{"test.com"}},
		AutoApproveEmployees: true,
	}
	do := toCorpSigningDOForMigrate(cs)
	if !do.AutoApproveEmployees {
		t.Error("toCorpSigningDOForMigrate: AutoApproveEmployees should be true")
	}
}
