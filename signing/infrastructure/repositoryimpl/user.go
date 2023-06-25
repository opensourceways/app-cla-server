package repositoryimpl

import (
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
)

func NewUser(dao dao) *user {
	return &user{
		dao: dao,
	}
}

type user struct {
	dao dao
}

func (impl *user) Add(*domain.User) error {
	return nil
}

func (impl *user) Remove(dp.Account) error {
	return nil
}

func (impl *user) Save(*domain.User) error {
	return nil
}
func (impl *user) FindByAccount(dp.Account, string) (domain.User, error) {
	return domain.User{}, nil
}
func (impl *user) FindByEmail(dp.EmailAddr, string) (domain.User, error) {
	return domain.User{}, nil
}
