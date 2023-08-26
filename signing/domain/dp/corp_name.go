package dp

import (
	"errors"
	"regexp"

	"github.com/opensourceways/app-cla-server/util"
)

var reCorpNameXSS = regexp.MustCompile(`[<>"'/]`)

func NewCorpName(v string) (CorpName, error) {
	err := errors.New("invalid corp name")

	if v == "" {
		return nil, err
	}

	if util.StrLen(v) > config.MaxLengthOfCorpName {
		return nil, err
	}

	if reCorpNameXSS.MatchString(v) {
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
