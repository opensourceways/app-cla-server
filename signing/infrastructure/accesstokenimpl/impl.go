package accesstokenimpl

import (
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
	"github.com/opensourceways/app-cla-server/common/infrastructure/redisdb"
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

func (impl *accessTokenImpl) Add(value []byte) (string, error) {
	key, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	err = impl.cli.Set(redisdb.Ctx(), key.String(), value, 0).Err()
	if err != nil {
		return "", err
	}

	return key.String(), nil
}

func (impl *accessTokenImpl) Find(key string) ([]byte, error) {
	val, err := impl.cli.Get(redisdb.Ctx(), key).Bytes()
	if err == redis.Nil {
		return nil, commonRepo.NewErrorResourceNotFound(err)
	} else if err != nil {
		return nil, err
	} else {
		return val, nil
	}
}

func (impl *accessTokenImpl) Delete(key string) error {
	return impl.cli.Expire(redisdb.Ctx(), key, impl.cfg.expire()).Err()
}
