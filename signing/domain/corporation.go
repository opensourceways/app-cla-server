package domain

import (
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewCorporation(name dp.CorpName, email dp.EmailAddr) Corporation {
	return Corporation{
		Name:               name,
		AllEmailDomains:    []string{email.Domain()},
		PrimaryEmailDomain: email.Domain(),
	}
}

type Corporation struct {
	Name               dp.CorpName
	AllEmailDomains    []string
	PrimaryEmailDomain string
}
