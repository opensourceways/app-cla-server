package mongodb

import (
	"errors"
	"strings"
	"time"
)

type Config struct {
	Conn    string `json:"conn"       required:"true"`
	DBName  string `json:"db"         required:"true"`
	CAFile  string `json:"ca_file"    required:"true"`
	Timeout int64  `json:"timeout"`
}

func (cfg *Config) SetDefault() {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10
	}
}

func (cfg *Config) Validate() error {
	if !strings.Contains(cfg.Conn, "ssl=true") {
		return errors.New("invalid mongodb conn")
	}

	return nil
}

func (cfg *Config) timeout() time.Duration {
	return time.Duration(cfg.Timeout) * time.Second
}
