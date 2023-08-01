package app

import (
	"strconv"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

type CmdToAddLink struct {
	Org       domain.OrgInfo
	Email     dp.EmailAddr
	CLAs      []CmdToAddCLA
	Submitter string
}

func (cmd *CmdToAddLink) toLink() domain.Link {
	v := make([]domain.CLA, len(cmd.CLAs))
	for i := range cmd.CLAs {
		v[i] = cmd.CLAs[i].toCLA()
		v[i].Id = strconv.Itoa(i)
	}

	return domain.Link{
		Type:      dp.LinkTypeCLA,
		Org:       cmd.Org,
		CLAs:      v,
		CLANum:    len(cmd.CLAs),
		Submitter: cmd.Submitter,
	}
}

type CmdToFindCLAs struct {
	LinkId string
	Type   dp.CLAType
}

type LinkCLADTO struct {
	CLA   CLADetailDTO
	Org   domain.OrgInfo
	Email domain.EmailInfo
}

type LinkDTO struct {
	Org   domain.OrgInfo
	Email domain.EmailInfo
}

type CmdToListLink struct {
	Orgs     []string
	Platform string
}

func (cmd *CmdToListLink) toOpt(t dp.LinkType) repository.FindLinksOpt {
	return repository.FindLinksOpt{
		Type:     t,
		Orgs:     cmd.Orgs,
		Platform: cmd.Platform,
	}
}
