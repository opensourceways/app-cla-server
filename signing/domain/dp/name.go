package dp

import (
	"errors"

	"github.com/opensourceways/app-cla-server/util"
)

func NewName(v string) (Name, error) {
	err := errors.New("invalid name")

	if v == "" {
		return nil, err
	}

	if util.StrLen(v) > config.MaxLengthOfName {
		return nil, err
	}

	if util.HasXSS(v) {
		return nil, err
	}

	return name(v), nil
}

// Name
type Name interface {
	Name() string
}

type name string

func (r name) Name() string {
	return string(r)
}
