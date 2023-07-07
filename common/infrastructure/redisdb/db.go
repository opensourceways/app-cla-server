package redisdb

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	cli     *redis.Client
	timeout time.Duration
)

func Init(cfg *Config) error {
	timeout = cfg.timeout()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return WithContext(func(ctx context.Context) error {
		_, err := rdb.Ping(ctx).Result()

		return err
	})
}

func Close() error {
	if cli != nil {
		return cli.Close()
	}

	return nil
}

func DAO() *daoImpl {
	return &daoImpl{
		instance: cli,
	}
}

func WithContext(f func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return f(ctx)
}
