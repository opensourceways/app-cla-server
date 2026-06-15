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

type IndividualSignedDTO struct {
	Type           string `json:"type"`
	Signed         bool   `json:"signed"`
	VersionMatched bool   `json:"version_matched"`
	PendingVersion bool   `json:"pending_version"`
}

type CmdToFindSignedCLAInfo = CmdToCheckSinging

type CLAInfoDTO struct {
	CLAId    string `json:"cla_id"`
	Language string `json:"language"`
}
