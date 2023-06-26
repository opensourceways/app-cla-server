package repository

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type User interface {
	Add(*domain.User) (string, error)
	Remove(string) error
	RemoveByAccount(dp.Account) error
	SavePassword(*domain.User) error
	Find(string) (domain.User, error)
}
