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
		Addr: cfg.Address,
		DB:   cfg.DB,
	})

	_, err := rdb.Ping(Ctx()).Result()
	if err != nil {
		return err
	}

	return nil
}

func Close() error {
	if cli != nil {
		return cli.Close()
	}

	return nil
}

func Instance() *redis.Client {
	return cli
}

func Ctx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), timeout)

	return ctx
}
