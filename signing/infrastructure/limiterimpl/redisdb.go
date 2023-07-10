package limiterimpl

import "time"

type dao interface {
	SetKey(key string, expiry time.Duration) error
	HasKey(key string) (bool, error)
}
