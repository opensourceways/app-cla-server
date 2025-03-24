package dp

import (
	"errors"
	"regexp"

	"github.com/opensourceways/app-cla-server/util"
)

var reAccount = regexp.MustCompile("^[a-zA-Z0-9_.-]+_[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z]{2,6}$")

func NewAccount(v string) (Account, error) {
	err := errors.New("invalid account")

	if util.StrLen(v) > config.MaxLengthOfAccount {
		return nil, err
	}

	if v == "" || !reAccount.MatchString(v) {
		return nil, err
	}

	return account(v), nil
}

func CreateAccount(v string) Account {
	return account(v)
}

// Account
type Account interface {
	Account() string
}

type account string

func (r account) Account() string {
	return string(r)
}
