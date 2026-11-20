package vcratelimitimpl

import "time"

// Config holds configuration for the verification-code IP rate limiter.
type Config struct {
	// Threshold is the maximum number of verification-code requests allowed
	// per client IP within the Window. Default 10.
	Threshold int64 `json:"threshold"`
	// Window is the rate-limiting time window in seconds. Default 60.
	Window int64 `json:"window"`
}

func (cfg *Config) SetDefault() {
	if cfg.Threshold <= 0 {
		cfg.Threshold = 10
	}

	if cfg.Window <= 0 {
		cfg.Window = 60
	}
}

func (cfg *Config) window() time.Duration {
	return time.Duration(cfg.Window) * time.Second
}
