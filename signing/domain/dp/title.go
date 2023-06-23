package dp

import (
	"errors"

	"github.com/opensourceways/app-cla-server/util"
)

func NewTitle(v string) (Title, error) {
	if v == "" {
		return nil, errors.New("invalid title")
	}

	if util.StrLen(v) > config.MaxLengthOfTitle {
		return nil, errors.New("invalid title")
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
