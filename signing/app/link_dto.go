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
	CLAs      []CLAInfo
	Submitter string
}

func (cmd *CmdToAddLink) toLink() domain.Link {
	v := make([]domain.CLA, len(cmd.CLAs))
	for i := range cmd.CLAs {
		v[i] = cmd.CLAs[i].toCLA()
		v[i].Id = strconv.Itoa(i)
	}

	return domain.Link{
		Org:       cmd.Org,
		CLAs:      v,
		CLANum:    len(cmd.CLAs),
		Submitter: cmd.Submitter,

		// GracePeriodDays 留空(nil)表示未显式配置，读取时自动回退到全局默认宽限期天数，
		// 不需要显式赋值，也不需要任何数据迁移。
	}
}

type CmdToListLink = repository.FindLinksOpt

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
