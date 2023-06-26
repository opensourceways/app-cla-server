package repositoryimpl

import (
	"go.mongodb.org/mongo-driver/bson"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func NewUser(dao dao) *user {
	return &user{
		dao: dao,
	}
}

type user struct {
	dao dao
}

func (impl *user) Add(v *domain.User) (string, error) {
	do := toUserDO(v)
	doc, err := do.toDoc()
	if err != nil {
		return "", err
	}
	doc[fieldVersion] = 0

	docFilter := linkIdFilter(v.LinkId)
	docFilter[mongodbCmdOr] = bson.A{
		bson.M{fieldEmail: v.EmailAddr.EmailAddr()},
		bson.M{fieldAccount: v.Account.Account()},
	}

	index, err := impl.dao.InsertDocIfNotExists(docFilter, doc)
	if err != nil && impl.dao.IsDocExists(err) {
		err = commonRepo.NewErrorDuplicateCreating(err)
	}

	return index, err
}

func (impl *user) Remove(index string) error {
	filter, err := impl.dao.DocIdFilter(index)
	if err != nil {
		return err
	}

	return impl.dao.DeleteDoc(filter)
}

func (impl *user) SavePassword(u *domain.User) error {
	filter, err := impl.dao.DocIdFilter(u.Id)
	if err != nil {
		return err
	}

	doc := bson.M{
		fieldPassword: u.Password.Password(),
		fieldChanged:  u.PasswordChaged,
	}

	err = impl.dao.UpdateDoc(filter, doc, u.Version)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}

func (impl *user) Find(string) (domain.User, error) {
	return domain.User{}, nil
}
