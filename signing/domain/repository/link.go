package repository

import "github.com/opensourceways/app-cla-server/signing/domain"

type FindLinksOpt struct {
	Platform string
	Orgs     []string
}

type LinkSummary struct {
	Id        string
	Org       domain.OrgInfo
	Email     domain.EmailInfo
	Submitter string
}

type LinkCLA struct {
	Id          string
	Org         domain.OrgInfo
	Email       domain.EmailInfo
	Clas        []domain.CLA
	RemovedCLAs []domain.CLA
	// nil 表示未显式配置，由调用方回退到全局默认宽限期天数
	GracePeriodDays *int
}

type Link interface {
	NewLinkId() string
	Add(*domain.Link) error
	Remove(*domain.Link) error
	Find(string) (domain.Link, error)
	FindAll(userId string) ([]LinkSummary, error)
	ListAll() ([]LinkCLA, error)

	AddCLA(*domain.Link, *domain.CLA) error
	UpdateCLA(link *domain.Link, newCla *domain.CLA) error
	RemoveCLA(*domain.Link, *domain.CLA) error
	UpdateGracePeriodDays(linkId string, days int) error
}
