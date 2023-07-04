package randomcodeimpl

import "crypto/rand"

const (
	codeLen    = 6
	dictionary = "0123456789"
)

func NewRandomCodeImpl() randomCodeImpl {
	return randomCodeImpl{}
}

type randomCodeImpl struct{}

func (impl randomCodeImpl) New() (string, error) {
	var bytes = make([]byte, codeLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	n := byte(len(dictionary))
	for k, v := range bytes {
		bytes[k] = dictionary[v%n]
	}

	return string(bytes), nil
}

func (impl randomCodeImpl) IsValid(s string) bool {
	if len(s) != codeLen {
		return false
	}

	for _, c := range s {
		if !(c >= '0' && c <= '9') {
			return false
		}
	}

	return true
}
