package loginimpl

import "time"

type dao interface {
	SetWithExpiry(key string, val interface{}, expiry time.Duration) error
	Get(key string, val interface{}) error
	Expire(key string, expire time.Duration) error
	IsDocNotExists(err error) bool
}
