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

func (impl *user) Add(v *domain.User) (string, error) {
	do := toUserDO(v)
	doc, err := do.toDoc()
	if err != nil {
		return "", err
	}
	doc[fieldVersion] = 0
	doc[fieldPassword] = v.Password

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

func (impl *user) Remove(ids []string) error {
	if len(ids) == 1 {
		filter, err := impl.dao.DocIdFilter(ids[0])
		if err != nil {
			return err
		}

		return impl.dao.DeleteDoc(filter)
	}

	filter, err := impl.dao.DocIdsFilter(ids)
	if err != nil {
		return err
	}

	return impl.dao.DeleteDocs(filter)
}

func (impl *user) RemoveByAccount(linkId string, accounts []dp.Account) error {
	filter := linkIdFilter(linkId)

	if len(accounts) == 1 {
		filter[fieldAccount] = accounts[0].Account()

		return impl.dao.DeleteDoc(filter)
	}

	v := make(bson.A, len(accounts))
	for i := range accounts {
		v[i] = accounts[i].Account()
	}
	filter[fieldAccount] = bson.M{mongodbCmdIn: v}

	return impl.dao.DeleteDocs(filter)
}

func (impl *user) SaveLoginInfo(u *domain.User) error {
	filter, err := impl.dao.DocIdFilter(u.Id)
	if err != nil {
		return err
	}

	doc := bson.M{
		fieldFailedNum:  u.FailedNum,
		fieldLoginTime:  u.LoginTime,
		fieldFrozenTime: u.FrozenTime,
	}

	err = impl.dao.UpdateDoc(filter, doc, u.Version)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}

func (impl *user) SavePassword(u *domain.User) error {
	filter, err := impl.dao.DocIdFilter(u.Id)
	if err != nil {
		return err
	}

	doc := bson.M{
		fieldPassword: u.Password,
		fieldChanged:  u.PasswordChaged,
	}

	err = impl.dao.UpdateDoc(filter, doc, u.Version)
	if err != nil && impl.dao.IsDocNotExists(err) {
		err = commonRepo.NewErrorConcurrentUpdating(err)
	}

	return err
}

func (impl *user) Find(index string) (u domain.User, err error) {
	filter, err := impl.dao.DocIdFilter(index)
	if err != nil {
		return
	}

	var do userDO

	if err = impl.dao.GetDoc(filter, nil, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}
	} else {
		err = do.toUser(&u)
	}

	return
}

func (impl *user) FindByAccount(linkId string, a dp.Account) (u domain.User, err error) {
	filter := linkIdFilter(linkId)
	filter[fieldAccount] = a.Account()

	var do userDO

	if err = impl.dao.GetDoc(filter, nil, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}
	} else {
		err = do.toUser(&u)
	}

	return
}

func (impl *user) FindByEmail(linkId string, e dp.EmailAddr) (u domain.User, err error) {
	filter := linkIdFilter(linkId)
	filter[fieldEmail] = e.EmailAddr()

	var do userDO

	if err = impl.dao.GetDoc(filter, nil, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}
	} else {
		err = do.toUser(&u)
	}

	return
}
