package redisdb

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
)

type daoImpl struct {
	instance *redis.Client
}

func (impl *daoImpl) Set(key string, val interface{}) error {
	return WithContext(func(ctx context.Context) error {
		return impl.instance.Set(ctx, key, val, 0).Err()
	})
}

func (impl *daoImpl) Get(key string, data interface{}) error {
	return WithContext(func(ctx context.Context) error {
		err := impl.instance.Get(ctx, key).Scan(data)
		if err == redis.Nil {
			return commonRepo.NewErrorResourceNotFound(err)
		}

		return err
	})
}

func (impl *daoImpl) Expire(key string, expire time.Duration) error {
	return WithContext(func(ctx context.Context) error {
		return impl.instance.Expire(ctx, key, expire).Err()
	})
}
