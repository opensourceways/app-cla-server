package app

import (
	"testing"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func TestCmdToSignIndividualCLAToIndividualSigning(t *testing.T) {
	e, _ := dp.NewEmailAddr("john@test.com")
	n := dp.CreateName("John")
	cmd := &CmdToSignIndividualCLA{
		Link: domain.LinkInfo{
			Id: "link-1",
			CLAInfo: domain.CLAInfo{
				CLAId:    "cla-1",
				Language: dp.CreateLanguage("en"),
			},
		},
		Rep: domain.Representative{
			Name:      n,
			EmailAddr: e,
		},
		AllSingingInfo:   domain.AllSingingInfo{"field1": "val1"},
		VerificationCode: "123456",
	}
	signing := cmd.toIndividualSigning()
	if signing.Link.Id != "link-1" || signing.Rep.Name.Name() != "John" {
		t.Errorf("toIndividualSigning mismatch")
	}
}

func TestCmdToSignIndividualCLAToCmd(t *testing.T) {
	e, _ := dp.NewEmailAddr("john@test.com")
	cmd := &CmdToSignIndividualCLA{
		Link: domain.LinkInfo{Id: "link-1"},
		Rep:  domain.Representative{EmailAddr: e},
	}
	vcCmd := cmd.toCmd()
	if vcCmd.Id != "link-1" || vcCmd.EmailAddr.EmailAddr() != "john@test.com" {
		t.Errorf("toCmd mismatch")
	}
}

func TestCmdToSignCorpCLAToCorpSigning(t *testing.T) {
	e, _ := dp.NewEmailAddr("ceo@corp.com")
	n := dp.CreateName("CEO")
	cn := dp.CreateCorpName("BigCorp")
	cmd := &CmdToSignCorpCLA{
		Link: domain.LinkInfo{
			Id: "link-1",
			CLAInfo: domain.CLAInfo{
				CLAId:    "cla-1",
				Language: dp.CreateLanguage("en"),
			},
		},
		CorpName: cn,
		Rep: domain.Representative{
			Name:      n,
			EmailAddr: e,
		},
		AllSingingInfo:   domain.AllSingingInfo{"f1": "v1"},
		VerificationCode: "654321",
	}
	cs := cmd.toCorpSigning()
	if cs.Corp.Name.CorpName() != "BigCorp" || cs.Rep.Name.Name() != "CEO" {
		t.Errorf("toCorpSigning mismatch")
	}
	if cs.Date == "" {
		t.Error("expected date to be set")
	}
}

func TestCmdToSignCorpCLAToCmd(t *testing.T) {
	e, _ := dp.NewEmailAddr("ceo@corp.com")
	cmd := &CmdToSignCorpCLA{
		Link: domain.LinkInfo{Id: "link-1"},
		Rep:  domain.Representative{EmailAddr: e},
	}
	vcCmd := cmd.toCmd()
	if vcCmd.Id != "link-1" || vcCmd.EmailAddr.EmailAddr() != "ceo@corp.com" {
		t.Errorf("toCmd mismatch")
	}
}

func TestCmdToFindCLAs(t *testing.T) {
	cmd := &CmdToFindCLAs{
		LinkId: "link-1",
		Type:   dp.CLATypeCorp,
	}
	if cmd.LinkId != "link-1" || cmd.Type.CLAType() != "corporation" {
		t.Errorf("CmdToFindCLAs mismatch")
	}
}

func TestCorpSigningDTO(t *testing.T) {
	dto := CorpSigningDTO{
		Id:             "cs-1",
		Date:           "2024-01-01",
		Language:       "en",
		CorpName:       "BigCorp",
		RepName:        "CEO",
		RepEmail:       "ceo@corp.com",
		HasAdminAdded:  true,
		HasPDFUploaded: false,
	}
	if dto.Id != "cs-1" {
		t.Errorf("dto mismatch")
	}
}

func TestCorpSigningInfoDTO(t *testing.T) {
	dto := CorpSigningInfoDTO{
		Date:     "2024-01-01",
		CLAId:    "cla-1",
		Language: "en",
		CorpName: "BigCorp",
		RepName:  "CEO",
		RepEmail: "ceo@corp.com",
		AllInfo:  domain.AllSingingInfo{"k": "v"},
	}
	if dto.CorpName != "BigCorp" {
		t.Errorf("dto mismatch")
	}
}

func TestCLADTO(t *testing.T) {
	dto := CLADTO{
		Id:       "cla-1",
		Type:     "corporation",
		URL:      "https://example.com/a.pdf",
		Language: "en",
	}
	if dto.Id != "cla-1" || dto.Type != "corporation" {
		t.Errorf("CLADTO mismatch")
	}
}

func TestCLADetailDTO(t *testing.T) {
	dto := CLADetailDTO{
		Id:        "cla-1",
		Fileds:    []domain.Field{},
		Language:  "en",
		LocalFile: "/tmp/cla.pdf",
	}
	if dto.Id != "cla-1" {
		t.Errorf("CLADetailDTO mismatch")
	}
}

func TestLinkCLADTO(t *testing.T) {
	dto := LinkCLADTO{
		CLA: CLADetailDTO{Id: "cla-1"},
		Org: domain.OrgInfo{Alias: "Org"},
		Email: domain.EmailInfo{
			Addr:     dp.CreateEmailAddr("org@test.com"),
			Platform: "gmail",
		},
	}
	if dto.CLA.Id != "cla-1" || dto.Org.Alias != "Org" {
		t.Errorf("LinkCLADTO mismatch")
	}
}

func TestLinkDTO(t *testing.T) {
	dto := LinkDTO{
		Org:   domain.OrgInfo{Alias: "TestOrg"},
		Email: domain.EmailInfo{Platform: "smtp"},
	}
	if dto.Org.Alias != "TestOrg" {
		t.Errorf("LinkDTO mismatch")
	}
}

func TestCmdToFindCLAsStruct(t *testing.T) {
	_ = CmdToFindCLAs{LinkId: "L1", Type: dp.CLATypeIndividual}
	_ = CmdToCheckSinging{LinkId: "L1", EmailAddr: dp.CreateEmailAddr("t@t.com")}
	_ = CLAInfoDTO{CLAId: "c1", Language: "en"}
}
