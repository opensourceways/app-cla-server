package limiter

import "time"

type Limiter interface {
	Add(string, time.Duration) error
	IsAllowed(string) (bool, error)
}
