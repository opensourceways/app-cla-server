package dp

import (
	"errors"

	"github.com/opensourceways/app-cla-server/util"
)

func NewName(v string) (Name, error) {
	if v == "" {
		return nil, errors.New("invalid name")
	}

	if util.StrLen(v) > config.MaxLengthOfName {
		return nil, errors.New("invalid name")
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
