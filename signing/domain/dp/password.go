package dp

import "errors"

func NewPassword(v []byte) (Password, error) {
	if len(v) == 0 {
		return nil, errors.New("invalid password")
	}

	return password(v), nil
}

// Password
type Password interface {
	Password() []byte
	Clear()
}

type password []byte

func (r password) Password() []byte {
	return []byte(r)
}

func (r password) Clear() {
	for i := range r {
		r[i] = 0
	}
}
