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

func (c *Corporation) isMyEmail(email dp.EmailAddr) bool {
	domain := email.Domain()

	for _, v := range c.AllEmailDomains {
		if v == domain {
			return true
		}
	}

	return false
}
