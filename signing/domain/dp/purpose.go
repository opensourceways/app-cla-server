package dp

import "errors"

type Purpose interface {
	Purpose() string
}

type purpose string

func (v purpose) Purpose() string {
	return string(v)
}

func NewPurpose(v string) (Purpose, error) {
	if v == "" {
		return nil, errors.New("invalid purpose")
	}

	return purpose(v), nil
}

func CreatePurpose(v string) Purpose {
	return purpose(v)
}
