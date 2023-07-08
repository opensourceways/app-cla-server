package redisdb

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"

	commonRepo "github.com/opensourceways/app-cla-server/common/domain/repository"
)

func (cli *client) Set(key string, val interface{}) error {
	return cli.withContext(func(ctx context.Context) error {
		return cli.redisCli.Set(ctx, key, val, 0).Err()
	})
}

func (cli *client) Get(key string, data interface{}) error {
	return cli.withContext(func(ctx context.Context) error {
		err := cli.redisCli.Get(ctx, key).Scan(data)
		if err == redis.Nil {
			return commonRepo.NewErrorResourceNotFound(err)
		}

		return err
	})
}

func (cli *client) Expire(key string, expire time.Duration) error {
	return cli.withContext(func(ctx context.Context) error {
		return cli.redisCli.Expire(ctx, key, expire).Err()
	})
}
