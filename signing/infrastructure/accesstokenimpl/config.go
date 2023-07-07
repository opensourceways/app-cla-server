package accesstokenimpl

import "time"

type Config struct {
	// unit second
	Expire int64 `json:"expire"`
}

func (cfg *Config) SetDefault() {
	if cfg.Expire <= 0 {
		cfg.Expire = 5
	}
}

func (cfg *Config) expire() time.Duration {
	return time.Duration(cfg.Expire) * time.Second
}
