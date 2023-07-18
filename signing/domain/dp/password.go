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

func IsSamePassword(p1, p2 Password) bool {
	v1 := p1.Password()
	v2 := p2.Password()

	if len(v1) != len(v2) {
		return false
	}

	for i := range v1 {
		if v1[i] != v2[i] {
			return false
		}
	}

	return true
}
