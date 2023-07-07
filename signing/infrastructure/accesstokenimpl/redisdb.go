package accesstokenimpl

import "time"

type dao interface {
	Set(key string, val interface{}) error
	Get(key string, val interface{}) error
	Expire(key string, expire time.Duration) error
}
