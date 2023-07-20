package loginimpl

import (
	"time"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func NewLoginImpl(d dao, c *Config) *loginImpl {
	return &loginImpl{
		dao:    d,
		expiry: c.expire(),
	}
}

type loginImpl struct {
	dao dao

	expiry time.Duration
}

func (impl *loginImpl) Add(value *domain.Login) error {
	v := toLoginDo(value)

	// must pass *loginDO, because it implements the interface of MarshalBinary
	return impl.dao.SetWithExpiry(value.Id, &v, impl.expiry)
}

func (impl *loginImpl) Find(key string) (domain.Login, error) {
	var do loginDO

	// note the *loginDO implements interface of UnmarshalBinary
	if err := impl.dao.Get(key, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}

		return domain.Login{}, err
	}

	return do.toLogin(key), nil
}

func (impl *loginImpl) Delete(key string) error {
	return impl.dao.Expire(key, 0)
}
