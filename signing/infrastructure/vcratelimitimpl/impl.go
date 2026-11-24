package vcratelimitimpl

import "time"

const vcRateLimitKeyPrefix = "vc:rate:"

func NewVCRateLimiter(d dao, cfg *Config) *vcRateLimiterImpl {
	return &vcRateLimiterImpl{
		dao:       d,
		threshold: cfg.Threshold,
		window:    cfg.window(),
	}
}

type vcRateLimiterImpl struct {
	dao       dao
	threshold int64
	window    time.Duration
}

// Allow increments the request counter for the given IP and returns true
// when the request is within the allowed quota. The TTL is set only on the
// first request of each window (fixed window).
func (l *vcRateLimiterImpl) Allow(ip string) (bool, error) {
	key := vcRateLimitKeyPrefix + ip

	count, err := l.dao.Incr(key)
	if err != nil {
		return false, err
	}

	if count == 1 {
		if err := l.dao.Expire(key, l.window); err != nil {
			return false, err
		}
	}

	return count <= l.threshold, nil
}
