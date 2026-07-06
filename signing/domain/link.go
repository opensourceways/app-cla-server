package domain

import (
	"strconv"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type EmailInfo struct {
	Addr     dp.EmailAddr
	Platform string
}

type OrgInfo struct {
	Alias      string // normal community name
	Logo       string
	ProjectURL string
}

type Link struct {
	Id        string
	Org       OrgInfo
	Email     EmailInfo
	CLAs      []CLA
	Submitter string
	CLANum    int
	Version   int

	// nil 表示未显式配置，回退全局默认宽限期天数；非nil则是管理员显式设置的值（含0）。
	GracePeriodDays *int
}

func (link *Link) CanDo(userId string) error {
	if userId != link.Submitter {
		return NewDomainError(ErrorCodeNoPermission)
	}

	return nil
}

func (link *Link) AddCLA(cla *CLA) error {
	if _, ok := link.posOfCLA(cla); ok {
		return NewDomainError(ErrorCodeCLAExists)
	}

	cla.Id = strconv.Itoa(link.CLANum)
	link.CLANum += 1

	return nil
}

func (link *Link) UpdateCLA(cla *CLA) error {
	if _, ok := link.posOfCLA(cla); !ok {
		return NewDomainError(ErrorCodeCLANotExists)
	}

	cla.Id = strconv.Itoa(link.CLANum)
	link.CLANum += 1

	return nil
}

func (link *Link) FindCLA(index string) *CLA {
	for i := range link.CLAs {
		if link.CLAs[i].Id == index {
			return &link.CLAs[i]
		}
	}

	return nil
}

func (link *Link) posOfCLA(cla *CLA) (int, bool) {
	for i := range link.CLAs {
		if link.CLAs[i].isMe(cla) {
			return i, true
		}
	}

	return 0, false
}

func (link *Link) GetCLA(t dp.CLAType, l dp.Language) *CLA {
	for i, v := range link.CLAs {
		if v.Type == t && v.Language == l {
			return &link.CLAs[i]
		}
	}

	return nil
}

func (link *Link) GetEffectiveGracePeriodDays(defaultDays int) int {
	// nil：字段缺失/从未配置过（新建link、或功能上线前的历史数据）
	// 负数：显式重置为“使用默认值”
	// 二者都回退到全局默认宽限期天数，不需要任何数据迁移。
	if link.GracePeriodDays == nil || *link.GracePeriodDays < 0 {
		return defaultDays
	}
	return *link.GracePeriodDays
}
