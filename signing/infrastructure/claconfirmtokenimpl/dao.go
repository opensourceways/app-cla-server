package claconfirmtokenimpl

import "time"

type dao interface {
	SetWithExpiry(key string, val interface{}, expiry time.Duration) error
	GetAndDelete(key string, val interface{}) error
	Get(key string, val interface{}) error
	Del(keys ...string) error
	IsDocNotExists(err error) bool
}
