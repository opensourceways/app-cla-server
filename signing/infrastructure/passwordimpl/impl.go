package passwordimpl

import (
	"crypto/rand"
	"regexp"
)

var pwRe = regexp.MustCompile("^[\x21-\x7E]+$")

func NewPasswordImpl(cfg *Config) *passwordImpl {
	return &passwordImpl{cfg: *cfg}
}

type passwordImpl struct {
	cfg Config
}

func (impl *passwordImpl) New() (string, error) {
	var bytes = make([]byte, impl.cfg.MinLength)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	n := byte(0x7E - 0x20)
	items := make([]byte, len(bytes))
	for k, v := range bytes {
		items[k] = 0x21 + v%n
	}

	if s := string(items); impl.goodFormat(s) {
		return s, nil
	}

	items[0] = byte('a') + bytes[0]%26
	items[1] = byte('A') + bytes[1]%26
	items[2] = byte('0') + bytes[1]%10

	return string(items), nil
}

func (impl *passwordImpl) IsValid(s string) bool {
	if n := len(s); n < impl.cfg.MinLength || n > impl.cfg.MaxLength {
		return false
	}

	if !pwRe.MatchString(s) {
		return false
	}

	return impl.goodFormat(s)
}

func (impl *passwordImpl) goodFormat(s string) bool {
	part := make([]bool, 4)

	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			part[0] = true
		} else if c >= 'A' && c <= 'Z' {
			part[1] = true
		} else if c >= '0' && c <= '9' {
			part[2] = true
		} else {
			part[3] = true
		}
	}

	i := 0
	for _, b := range part {
		if b {
			i++
		}
	}

	return i >= 3
}
