package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/util"
)

type CmdToSignEmployeeCLA struct {
	CLA            domain.CLA
	Rep            domain.Representative
	CorpSigningId  string
	AllSingingInfo domain.AllSingingInfo
}

func (cmd *CmdToSignEmployeeCLA) toEmployeeSigning() domain.EmployeeSigning {
	return domain.EmployeeSigning{
		CLA:     cmd.CLA,
		Rep:     cmd.Rep,
		Date:    util.Date(),
		AllInfo: cmd.AllSingingInfo,
	}
}
