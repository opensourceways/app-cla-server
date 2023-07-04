package accesstokenimpl

import (
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/common/infrastructure/redisdb"
	"github.com/opensourceways/app-cla-server/signing/domain"
)

func NewAccessTokenImpl(c *redis.Client, cfg *Config) *accessTokenImpl {
	return &accessTokenImpl{
		cli: c,
		cfg: cfg,
	}
}

type accessTokenImpl struct {
	cli *redis.Client
	cfg *Config
}

func (impl *accessTokenImpl) Add(value *domain.AccessToken) (string, error) {
	key, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	// 0 means never expire
	err = impl.cli.Set(redisdb.Ctx(), key.String(), value, 0).Err()
	if err != nil {
		return "", err
	}

	return key.String(), nil
}

func (impl *accessTokenImpl) Find(key string) (token domain.AccessToken, err error) {
	err = impl.cli.Get(redisdb.Ctx(), key).Scan(&token)
	if err == redis.Nil {
		return token, commonRepo.NewErrorResourceNotFound(err)
	} else if err != nil {
		return
	} else {
		return token, nil
	}
}

func (impl *accessTokenImpl) Delete(key string) error {
	return impl.cli.Expire(redisdb.Ctx(), key, impl.cfg.expire()).Err()
}
