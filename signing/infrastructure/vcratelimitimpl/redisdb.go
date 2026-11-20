package vcratelimitimpl

import "time"

type dao interface {
	Incr(key string) (int64, error)
	Expire(key string, expire time.Duration) error
}
