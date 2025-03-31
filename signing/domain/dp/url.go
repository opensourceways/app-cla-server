package dp

import (
	"errors"
	"net/url"
)

// URL
type URL interface {
	URL() string
}

func NewURL(v string) (URL, error) {
	if v == "" {
		return nil, errors.New("empty url")
	}

	if _, err := url.ParseRequestURI(v); err != nil {
		return nil, errors.New("invalid url")
	}

	return dpURL(v), nil
}

func CreateURL(v string) URL {
	return dpURL(v)
}

type dpURL string

func (v dpURL) URL() string {
	return string(v)
}
