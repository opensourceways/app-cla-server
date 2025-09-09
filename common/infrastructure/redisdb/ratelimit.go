package redisdb

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type IRateLimiter interface {
	CheckAndRecordRateLimit(key string) (bool, error)
}

// RateLimitConfig 频率限制配置
type RateLimiterConfig struct {
	KeyPrefix  string        // key前缀
	LimitCount int           // 限制次数
	TimeWindow time.Duration // 时间窗口
}

type RedisRateLimiter struct {
	config RateLimiterConfig
}

func NewRedisRateLimiter(config RateLimiterConfig) *RedisRateLimiter {
	return &RedisRateLimiter{config: config}
}

// CheckAndRecordRateLimit 检查并记录频率限制（原子操作）
func (r *RedisRateLimiter) CheckAndRecordRateLimit(key string) (bool, error) {
	cli := DAO()
	if cli == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	var allowed bool
	err := cli.withContext(func(ctx context.Context) error {
		fullKey := fmt.Sprintf("%s:%s", r.config.KeyPrefix, key)

		// 使用Lua脚本保证原子性
		luaScript := redis.NewScript(`
            local key = KEYS[1]
            local limit = tonumber(ARGV[1])
            local expire = tonumber(ARGV[2])
            
            local current = redis.call("INCR", key)
            if current == 1 then
				redis.call("EXPIRE", key, expire)
            end

            if current > limit then
                return 0
            end

            return 1
        `)

		result, err := luaScript.Run(ctx, cli.redisCli, []string{fullKey},
			r.config.LimitCount, int(r.config.TimeWindow.Seconds())).Result()
		if err != nil {
			return err
		}

		allowed = result.(int64) == 1
		return nil
	})

	return allowed, err
}
