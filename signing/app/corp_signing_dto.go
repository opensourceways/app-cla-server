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
		Logs: []domain.CorpSigningLog{
			{
				Date:   util.Date(),
				CLAId:  cmd.Link.CLAId,
				Action: "sign",
			},
		},
	}
}

func (cmd *CmdToSignCorpCLA) toCmd() cmdToCreateCodeForCorpSigning {
	return cmdToCreateCodeForCorpSigning{
		Id:        cmd.Link.Id,
		EmailAddr: cmd.Rep.EmailAddr,
	}
}

type CorpSigningPageDTO struct {
	Total int64
	Data  []CorpSigningDTO
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
	Date         string                `json:"date"`
	CLAId        string                `json:"cla_id"`
	Language     string                `json:"cla_language"`
	CorpName     string                `json:"corporation_name"`
	RepName      string                `json:"rep_name"`
	RepEmail     string                `json:"rep_email"`
	AllInfo      domain.AllSingingInfo `json:"info"`
	PendingCLAId string                `json:"pending_cla_id"`
	Logs         []CorpSigningLogDTO   `json:"logs"`
}

type CorpSigningLogDTO struct {
	Date   string `json:"date"`
	CLAId  string `json:"cla_id"`
	Action string `json:"action"`
}

func toCorpSigningLogDTOs(logs []domain.CorpSigningLog) []CorpSigningLogDTO {
	if len(logs) == 0 {
		return nil
	}
	result := make([]CorpSigningLogDTO, len(logs))
	for i, v := range logs {
		result[i] = CorpSigningLogDTO{
			Date:   v.Date,
			CLAId:  v.CLAId,
			Action: v.Action,
		}
	}
	return result
}

type CmdToFindCorpSummary = CmdToCheckSinging

type CorpSummaryDTO struct {
	CorpName      string `json:"corp_name"`
	CorpSigningId string `json:"corp_signing_id"`
}
