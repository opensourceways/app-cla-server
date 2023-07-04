package randombytesimpl

import "crypto/rand"

func NewRandomBytesImpl() *randomBytesImpl {
	return &randomBytesImpl{}
}

type randomBytesImpl struct{}

func (impl *randomBytesImpl) New(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)

	return b, err
}
