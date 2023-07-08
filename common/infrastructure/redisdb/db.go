package redisdb

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

var cli *client

func Init(cfg *Config) error {
	timeout := cfg.timeout()

	v := redis.NewClient(&redis.Options{
		DB:       cfg.DB,
		Addr:     cfg.Address,
		Password: cfg.Password,
	})

	err := withContext(
		func(ctx context.Context) error {
			_, err := v.Ping(ctx).Result()

			return err
		},
		timeout,
	)
	if err != nil {
		return err
	}

	cli = &client{
		redisCli: v,
		timeout:  timeout,
	}

	return nil
}

func Close() error {
	if cli != nil {
		return cli.redisCli.Close()
	}

	return nil
}

func DAO() *client {
	return cli
}

func withContext(f func(context.Context) error, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return f(ctx)
}

// client
type client struct {
	redisCli *redis.Client
	timeout  time.Duration
}

func (cli *client) withContext(f func(context.Context) error) error {
	return withContext(f, cli.timeout)
}
