package passwordimpl

import (
	"crypto/rand"
	"regexp"

	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

const (
	charNum        = 26
	digitalNum     = 10
	firstCharOfPw  = 0x21
	firstLowercase = 'a'
	lastLowercase  = 'z'
	firstUppercase = 'A'
	lastUppercase  = 'Z'
	firstDigital   = '0'
	lastDigital    = '9'
)

var (
	pwRe    = regexp.MustCompile("^[\x21-\x7E]+$")
	pwRange = byte(0x7E - 0x20)
)

func NewPasswordImpl(cfg *Config) *passwordImpl {
	return &passwordImpl{cfg: *cfg}
}

type passwordImpl struct {
	cfg Config
}

func (impl *passwordImpl) New() (dp.Password, error) {
	var bytes = make([]byte, impl.cfg.MinLength)

	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	items := make([]byte, len(bytes))
	for k, v := range bytes {
		items[k] = firstCharOfPw + v%pwRange
	}

	if impl.goodFormat(items) {
		return dp.NewPassword(items)
	}

	for i := range bytes {
		items[i] = impl.genChar(i, bytes[i])
	}

	return dp.NewPassword(items)
}

func (impl *passwordImpl) genChar(i int, v byte) byte {
	switch i % 3 {
	case 0:
		return byte(firstLowercase) + v%charNum
	case 1:
		return byte(firstUppercase) + v%charNum
	default:
		return byte(firstDigital) + v%digitalNum
	}
}

func (impl *passwordImpl) IsValid(p dp.Password) bool {
	s := p.Password()

	if n := len(s); n < impl.cfg.MinLength || n > impl.cfg.MaxLength {
		return false
	}

	if !pwRe.Match(s) {
		return false
	}

	return impl.goodFormat(s)
}

func (impl *passwordImpl) goodFormat(s []byte) bool {
	return impl.hasMultiChars(s) && !impl.hasConsecutive(s)
}

func (impl *passwordImpl) hasMultiChars(s []byte) bool {
	part := make([]bool, 4)

	for _, c := range s {
		if c >= firstLowercase && c <= lastLowercase {
			part[0] = true
		} else if c >= firstUppercase && c <= lastUppercase {
			part[1] = true
		} else if c >= firstDigital && c <= lastDigital {
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

	return i >= impl.cfg.MinNumOfKindOfPasswordChar
}

func (impl *passwordImpl) hasConsecutive(str []byte) bool {
	max := impl.cfg.MaxNumOfConsecutiveChars

	count := 1
	for i := 1; i < len(str); i++ {
		if str[i] == str[i-1] {
			if count++; count > max {
				return true
			}
		} else {
			count = 1
		}
	}

	return false
}
