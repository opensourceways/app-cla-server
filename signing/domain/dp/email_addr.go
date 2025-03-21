package dp

import (
	"errors"

	"github.com/opensourceways/app-cla-server/util"
)

func NewEmailAddr(v string) (EmailAddr, error) {
	if util.StrLen(v) > config.MaxLengthOfEmail {
		return nil, errors.New("invalid email address")
	}

	if err := util.CheckEmail(v); err != nil {
		return nil, err
	}

	return emailAddr(v), nil
}

// EmailAddr
type EmailAddr interface {
	EmailAddr() string
	Domain() string
}

type emailAddr string

func (r emailAddr) EmailAddr() string {
	return string(r)
}

func (r emailAddr) Domain() string {
	return util.EmailSuffix(string(r))
}
