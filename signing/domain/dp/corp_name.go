package dp

import (
	"errors"

	"github.com/opensourceways/app-cla-server/util"
)

func NewCorpName(v string) (CorpName, error) {
	err := errors.New("invalid corp name")

	if v == "" {
		return nil, err
	}

	if util.StrLen(v) > config.MaxLengthOfCorpName {
		return nil, err
	}

	if util.HasXSS(v) {
		return nil, err
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
