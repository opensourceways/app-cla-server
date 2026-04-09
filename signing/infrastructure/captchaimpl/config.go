package captchaimpl

import "time"

// Config holds captcha storage configuration.
type Config struct {
	// Expiry is how long a captcha remains valid, in seconds.
	Expiry int64 `json:"expiry"`
}

func (cfg *Config) SetDefault() {
	if cfg.Expiry <= 0 {
		cfg.Expiry = 300 // 5 minutes default
	}
}

func (cfg *Config) expiry() time.Duration {
	return time.Duration(cfg.Expiry) * time.Second
}
