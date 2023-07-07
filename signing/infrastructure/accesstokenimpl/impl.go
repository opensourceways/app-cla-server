package accesstokenimpl

import (
	"github.com/google/uuid"

	"github.com/opensourceways/app-cla-server/signing/domain"
)

func NewAccessTokenImpl(d dao, cfg *Config) *accessTokenImpl {
	return &accessTokenImpl{
		dao: d,
		cfg: cfg,
	}
}

type accessTokenImpl struct {
	dao dao
	cfg *Config
}

func (impl *accessTokenImpl) Add(value *domain.AccessToken) (string, error) {
	key, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	err = impl.dao.Set(key.String(), toAccessTokenDo(value))
	if err != nil {
		return "", err
	}

	return key.String(), nil
}

func (impl *accessTokenImpl) Find(key string) (domain.AccessToken, error) {
	var do AccessTokenDO
	if err := impl.dao.Get(key, &do); err != nil {
		return domain.AccessToken{}, err
	}

	return do.toAccessToken(), nil
}

func (impl *accessTokenImpl) Delete(key string) error {
	return impl.dao.Expire(key, impl.cfg.expire())
}
