package app

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/util"
)

type CmdToSignCorpCLA struct {
	Link           domain.Link
	CorpName       dp.CorpName
	Representative domain.Representative
	AllSingingInfo domain.AllSingingInfo
}

func (cmd *CmdToSignCorpCLA) toCorpSigning() domain.CorpSigning {
	return domain.CorpSigning{
		Date:    util.Date(),
		Link:    cmd.Link,
		Rep:     cmd.Representative,
		Corp:    domain.NewCorporation(cmd.CorpName, cmd.Representative.EmailAddr),
		AllInfo: cmd.AllSingingInfo,
	}
}
