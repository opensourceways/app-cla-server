package dp

import "errors"

func NewLanguage(v string) (Language, error) {
	if v = config.getLanguage(v); v == "" {
		return nil, errors.New("invalid language")
	}

	return language(v), nil
}

type Language interface {
	Language() string
}

type language string

func (v language) Language() string {
	return string(v)
}
