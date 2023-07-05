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

func (impl *accessTokenImpl) Add(value *domain.AccessTokenDO) (string, error) {
	key, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	// 0 means never expire
	err = impl.dao.Set(key.String(), value, 0)
	if err != nil {
		return "", err
	}

	return key.String(), nil
}

func (impl *accessTokenImpl) Find(key string) (token domain.AccessTokenDO, err error) {
	err = impl.dao.Get(key, &token)

	return
}

func (impl *accessTokenImpl) Delete(key string) error {
	return impl.dao.Expire(key, impl.cfg.expire())
}
