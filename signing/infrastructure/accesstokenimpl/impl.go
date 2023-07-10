package accesstokenimpl

import (
	"time"

	"github.com/google/uuid"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func NewAccessTokenImpl(d dao, c *Config) *accessTokenImpl {
	return &accessTokenImpl{
		dao:    d,
		expiry: c.expire(),
	}
}

type accessTokenImpl struct {
	dao    dao
	expiry time.Duration
}

func (impl *accessTokenImpl) Add(value *domain.AccessToken) (string, error) {
	key, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	v := toAccessTokenDo(value)
	// must pass *accessTokenDO, because it implements the interface of MarshalBinary
	err = impl.dao.Set(key.String(), &v)
	if err != nil {
		return "", err
	}

	return key.String(), nil
}

func (impl *accessTokenImpl) Find(key string) (domain.AccessToken, error) {
	var do accessTokenDO

	// note the *accessTokenDO implements interface of UnmarshalBinary
	if err := impl.dao.Get(key, &do); err != nil {
		if impl.dao.IsDocNotExists(err) {
			err = commonRepo.NewErrorResourceNotFound(err)
		}

		return domain.AccessToken{}, err
	}

	return do.toAccessToken(), nil
}

func (impl *accessTokenImpl) Delete(key string) error {
	return impl.dao.Expire(key, impl.expiry)
}
