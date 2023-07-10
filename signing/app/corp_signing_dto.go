package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/util"
)

type CmdToSignCorpCLA struct {
	Link             domain.LinkInfo
	CorpName         dp.CorpName
	Rep              domain.Representative
	AllSingingInfo   domain.AllSingingInfo
	VerificationCode string
}

func (cmd *CmdToSignCorpCLA) toCorpSigning() domain.CorpSigning {
	return domain.CorpSigning{
		Date:    util.Date(),
		Link:    cmd.Link,
		Rep:     cmd.Rep,
		Corp:    domain.NewCorporation(cmd.CorpName, cmd.Rep.EmailAddr),
		AllInfo: cmd.AllSingingInfo,
	}
}

func (cmd *CmdToSignCorpCLA) toCmd() cmdToCreateCodeForCorpSigning {
	return cmdToCreateCodeForCorpSigning{
		Id:        cmd.Link.Id,
		EmailAddr: cmd.Rep.EmailAddr,
	}
}

type CorpSigningDTO struct {
	Id             string `json:"id"`
	Date           string `json:"date"`
	Language       string `json:"cla_language"`
	CorpName       string `json:"corporation_name"`
	RepName        string `json:"rep_name"`
	RepEmail       string `json:"rep_email"`
	HasAdminAdded  bool   `json:"has_admin_added"`
	HasPDFUploaded bool   `json:"has_pdf_uploaded"`
}

type CorpSigningInfoDTO struct {
	Date     string                `json:"date"`
	CLAId    string                `json:"cla_id"`
	Language string                `json:"cla_language"`
	CorpName string                `json:"corporation_name"`
	RepName  string                `json:"rep_name"`
	RepEmail string                `json:"rep_email"`
	AllInfo  domain.AllSingingInfo `json:"info"`
}

type CmdToFindCorpSummary = CmdToCheckSinging

type CorpSummaryDTO struct {
	CorpName      string `json:"corp_name"`
	CorpSigningId string `json:"corp_signing_id"`
}
