package dp

import (
	"errors"
	"fmt"
)

type Purpose interface {
	Purpose() string
}

type purpose string

func (v purpose) Purpose() string {
	return string(v)
}

func NewPurpose(v string) (Purpose, error) {
	if v == "" {
		return nil, errors.New("invalid purpose")
	}

	return purpose(v), nil
}

func NewPurposeOfSigning(linkId string, email EmailAddr) Purpose {
	return purpose(
		fmt.Sprintf("sign %s, %s", linkId, email.EmailAddr()),
	)
}

func NewPurposeOfAddingEmailDomain(csId string, email EmailAddr) Purpose {
	return purpose(
		fmt.Sprintf("add email domain: %s, %s", csId, email.EmailAddr()),
	)
}
