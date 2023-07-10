package domain

type Limiter struct {
	Key      string
	Expiry   int64 // seconds
	Interval int
}
