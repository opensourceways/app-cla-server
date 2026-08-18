package redisdb

import (
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

var errDocNotExists = errors.New("doc doesn't exist")

// getAndDeleteScript atomically reads and deletes a key, so that a one-time
// value can be consumed exactly once even under concurrent requests.
// It returns 0 when the key doesn't exist, otherwise the stored value.
var getAndDeleteScript = redis.NewScript(`
local v = redis.call("GET", KEYS[1])
if not v then
	return 0
end
redis.call("DEL", KEYS[1])
return v
`)

func (cli *client) Set(key string, val interface{}) error {
	return cli.withContext(func(ctx context.Context) error {
		return cli.redisCli.Set(ctx, key, val, 0).Err()
	})
}

func (cli *client) SetWithExpiry(key string, val interface{}, expiry time.Duration) error {
	return cli.withContext(func(ctx context.Context) error {
		return cli.redisCli.Set(ctx, key, val, expiry).Err()
	})
}

func (cli *client) Get(key string, data interface{}) error {
	return cli.withContext(func(ctx context.Context) error {
		err := cli.redisCli.Get(ctx, key).Scan(data)
		if err == redis.Nil {
			return errDocNotExists
		}

		return err
	})
}

// GetAndDelete atomically fetches the value of the key and deletes the key.
// It returns errDocNotExists if the key doesn't exist.
func (cli *client) GetAndDelete(key string, data interface{}) error {
	return cli.withContext(func(ctx context.Context) error {
		res, err := getAndDeleteScript.Run(ctx, cli.redisCli, []string{key}).Result()
		if err != nil {
			return err
		}

		if s, ok := res.(string); ok && s != "" {
			return unmarshalTo(s, data)
		}

		return errDocNotExists
	})
}

func unmarshalTo(s string, data interface{}) error {
	if u, ok := data.(encoding.BinaryUnmarshaler); ok {
		return u.UnmarshalBinary([]byte(s))
	}

	return json.Unmarshal([]byte(s), data)
}

func (cli *client) Expire(key string, expire time.Duration) error {
	return cli.withContext(func(ctx context.Context) error {
		return cli.redisCli.Expire(ctx, key, expire).Err()
	})
}

func (cli *client) SetKey(key string, expiry time.Duration) error {
	return cli.withContext(func(ctx context.Context) error {
		return cli.redisCli.Set(ctx, key, 0, expiry).Err()
	})
}

func (cli *client) HasKey(key string) (bool, error) {
	exists := false
	err := cli.withContext(func(ctx context.Context) error {
		n, err := cli.redisCli.Exists(ctx, key).Result()
		if err != nil {
			return err
		}

		exists = n > 0

		return nil
	})

	return exists, err
}

func (impl *client) IsDocNotExists(err error) bool {
	return errors.Is(err, errDocNotExists)
}

func (cli *client) Del(keys ...string) error {
	return cli.withContext(func(ctx context.Context) error {
		return cli.redisCli.Del(ctx, keys...).Err()
	})
}

func (cli *client) Keys(pattern string) ([]string, error) {
	var result []string
	err := cli.withContext(func(ctx context.Context) error {
		iter := cli.redisCli.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			result = append(result, iter.Val())
		}
		return iter.Err()
	})
	return result, err
}
