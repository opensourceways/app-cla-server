package dp

import (
	"errors"

	"github.com/opensourceways/app-cla-server/util"
)

func NewTitle(v string) (Title, error) {
	err := errors.New("invalid title")

	if v == "" {
		return nil, err
	}

	if util.StrLen(v) > config.MaxLengthOfTitle {
		return nil, err
	}

	if util.HasXSS(v) {
		return nil, err
	}

	return title(v), nil
}

// Title
type Title interface {
	Title() string
}

type title string

func (r title) Title() string {
	return string(r)
}
