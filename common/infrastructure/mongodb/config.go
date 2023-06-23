package mongodb

import "time"

type Config struct {
	Conn    string `json:"conn"       required:"true"`
	DBName  string `json:"db"         required:"true"`
	Timeout int64  `json:"timeout"`
}

func (cfg *Config) SetDefault() {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10
	}
}

func (cfg *Config) timeout() time.Duration {
	return time.Duration(cfg.Timeout) * time.Second
}
