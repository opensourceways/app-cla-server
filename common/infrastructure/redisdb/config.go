package redisdb

import "time"

type Config struct {
	Address  string `json:"address"  required:"true"`
	Password string `json:"password" required:"true"`
	DB       int    `json:"db"`
	Timeout  int64  `json:"timeout"`
}

func (cfg *Config) SetDefault() {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 3
	}
}

func (cfg *Config) timeout() time.Duration {
	return time.Duration(cfg.Timeout) * time.Second
}
