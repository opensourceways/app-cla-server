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
	Id   string
	Clas []domain.CLA
}

type Link interface {
	NewLinkId() string
	Add(*domain.Link) error
	Remove(*domain.Link) error
	Find(string) (domain.Link, error)
	FindAll(userId string) ([]LinkSummary, error)
	ListAll() ([]LinkCLA, error)

	AddCLA(*domain.Link, *domain.CLA) error
	UpdateCLA(link *domain.Link, oldClaId string, newCla *domain.CLA) error
	RemoveCLA(*domain.Link, *domain.CLA) error
}
