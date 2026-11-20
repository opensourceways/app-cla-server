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
// when the request is within the allowed quota. On the first request in a
// window the key expiry is set so the counter resets automatically.
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
