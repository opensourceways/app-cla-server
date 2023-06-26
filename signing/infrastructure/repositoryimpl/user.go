package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
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

func (impl *user) Add(v *domain.User) error {
	do := toUserDO(v)
	doc, err := do.toDoc()
	if err != nil {
		return err
	}
	doc[fieldVersion] = 0

	docFilter := linkIdFilter(v.LinkId)
	docFilter[mongodbCmdOr] = bson.A{
		bson.M{fieldEmail: v.EmailAddr.EmailAddr()},
		bson.M{fieldAccount: v.Account.Account()},
	}

	_, err = impl.dao.InsertDocIfNotExists(docFilter, doc)
	if err != nil && impl.dao.IsDocExists(err) {
		err = commonRepo.NewErrorDuplicateCreating(err)
	}

	return err
}

func (impl *user) Remove(linkId string, a dp.Account) error {
	filter := linkIdFilter(linkId)
	filter[fieldAccount] = a.Account()

	return impl.dao.DeleteDoc(filter)
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
