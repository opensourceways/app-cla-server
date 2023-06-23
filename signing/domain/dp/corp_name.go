package dp

import (
	"errors"

	"github.com/opensourceways/app-cla-server/util"
)

// CorpName
type CorpName interface {
	CorpName() string
}

func NewCorpName(v string) (CorpName, error) {
	if v == "" {
		return nil, errors.New("invalid corp name")
	}

	if max := config.MaxLengthOfCorpName; util.StrLen(v) > max {
		return nil, errors.New("invalid corp name")
	}

	return corpName(v), nil
}

type corpName string

func (r corpName) CorpName() string {
	return string(r)
}
