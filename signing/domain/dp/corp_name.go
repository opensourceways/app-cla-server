package dp

import (
	"errors"

	"github.com/opensourceways/app-cla-server/util"
)

func NewCorpName(v string) (CorpName, error) {
	if v == "" {
		return nil, errors.New("invalid corp name")
	}

	if util.StrLen(v) > config.MaxLengthOfCorpName {
		return nil, errors.New("invalid corp name")
	}

	return corpName(v), nil
}

// CorpName
type CorpName interface {
	CorpName() string
}

type corpName string

func (r corpName) CorpName() string {
	return string(r)
}
