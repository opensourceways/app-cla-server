package app

import (
	"strconv"

	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/repository"
)

type CmdToAddDCOLink struct {
	Org       domain.OrgInfo
	Email     dp.EmailAddr
	DCOs      []CmdToAddDCO
	Submitter string
}

func (cmd *CmdToAddDCOLink) toDCOLink() domain.Link {
	v := make([]domain.CLA, len(cmd.DCOs))
	for i := range cmd.DCOs {
		v[i] = cmd.DCOs[i].toDCO()
		v[i].Id = strconv.Itoa(i)
	}

	return domain.Link{
		Org:       cmd.Org,
		CLAs:      v,
		CLANum:    len(cmd.DCOs),
		Submitter: cmd.Submitter,
	}
}

type CmdToListDCOLink = repository.FindLinksOpt
