package domain

import (
	"strings"

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

func (c *Corporation) isMyEmailIgnoreCase(email dp.EmailAddr) bool {
	domain := email.Domain()

	for _, v := range c.AllEmailDomains {
		if strings.EqualFold(v, domain) {
			return true
		}
	}

	return false
}

func (c *Corporation) addEmailDomain(ed string) error {
	for _, v := range c.AllEmailDomains {
		if v == ed {
			return NewDomainError(ErrorCodeCorpEmailDomainExists)
		}
	}

	if !c.isSubDomainOfPrimary(ed) {
		return NewDomainError(ErrorCodeCorpEmailDomainNotMatch)
	}

	c.AllEmailDomains = append(c.AllEmailDomains, ed)

	return nil
}

func (c *Corporation) isSubDomainOfPrimary(ed string) bool {
	primary := c.PrimaryEmailDomain
	if primary == "" {
		return false
	}

	return strings.HasSuffix(strings.ToLower(ed), "."+strings.ToLower(primary))
}
