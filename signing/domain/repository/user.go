package repository

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

type User interface {
	Add(*domain.User) (string, error)
	Remove(string) error
	SavePassword(*domain.User) error
	FindByAccount(dp.Account, string) (domain.User, error)
	FindByEmail(dp.EmailAddr, string) (domain.User, error)
}
