package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/util"
)

type CmdToSignIndividualCLA struct {
	Link             domain.LinkInfo
	Rep              domain.Representative
	AllSingingInfo   domain.AllSingingInfo
	VerificationCode string
}

func (cmd *CmdToSignIndividualCLA) toIndividualSigning() domain.IndividualSigning {
	return domain.IndividualSigning{
		Link:    cmd.Link,
		Rep:     cmd.Rep,
		Date:    util.Date(),
		AllInfo: cmd.AllSingingInfo,
	}
}

func (cmd *CmdToSignIndividualCLA) toCmd() cmdToCreateCodeForIndividualSigning {
	return cmdToCreateCodeForIndividualSigning{
		Id:        cmd.Link.Id,
		EmailAddr: cmd.Rep.EmailAddr,
	}
}

type CmdToCheckSinging struct {
	LinkId    string
	EmailAddr dp.EmailAddr
}
