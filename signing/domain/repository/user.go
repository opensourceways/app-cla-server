package repository

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type User interface {
	Add(*domain.User) (string, error)
	Remove([]string) error
	RemoveByAccount(linkId string, accounts []dp.Account) error
	SavePassword(*domain.User) error
	Find(string) (domain.User, error)
	FindByAccount(linkId string, a dp.Account) (domain.User, error)
	FindByEmail(linkId string, e dp.EmailAddr) (domain.User, error)
}
