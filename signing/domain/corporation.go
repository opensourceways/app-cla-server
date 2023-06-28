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

func (c *Corporation) addEmailDomain(ed string) error {
	for _, v := range c.AllEmailDomains {
		if v == ed {
			return NewDomainError(ErrorCodeCorpEmailDomainExists)
		}
	}

	if err := c.isValidEmailDomain(ed); err != nil {
		return err
	}

	c.AllEmailDomains = append(c.AllEmailDomains, ed)

	return nil
}

func (c *Corporation) isValidEmailDomain(ed string) error {
	e1 := strings.Split(c.PrimaryEmailDomain, ".")
	e2 := strings.Split(ed, ".")

	n1 := len(e1) - 1
	j := len(e2) - 1
	i := n1
	for ; i >= 0; i-- {
		if j < 0 {
			break
		}

		if e1[i] != e2[j] {
			break
		}

		j--
	}

	if i < 0 || n1-i >= config.MinNumOfSameEmailDomainParts {
		return nil
	}

	return NewDomainError(ErrorCodeCorpEmailDomainNotMatch)
}
