package dp

import "errors"

func NewPassword(v string) (Password, error) {
	if v == "" {
		return nil, errors.New("invalid password")
	}

	return password(v), nil
}

// Password
type Password interface {
	Password() string
}

type password string

func (r password) Password() string {
	return string(r)
}
