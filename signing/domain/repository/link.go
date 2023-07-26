package repository

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type FindLinksOpt struct {
	Type     dp.LinkType
	Orgs     []string
	Platform string
}

type LinkSummary struct {
	Id        string
	Org       domain.OrgInfo
	Email     domain.EmailInfo
	Submitter string
}

type Link interface {
	Add(*domain.Link) error
	Remove(*domain.Link) error
	Find(string) (domain.Link, error)
	FindAll(*FindLinksOpt) ([]LinkSummary, error)

	AddCLA(*domain.Link, *domain.CLA) error
	RemoveCLA(*domain.Link, *domain.CLA) error
}
