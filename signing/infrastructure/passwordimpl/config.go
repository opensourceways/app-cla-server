package passwordimpl

type Config struct {
	MinLength int `json:"min_length"`
	MaxLength int `json:"max_length"`
}

func (cfg *Config) SetDefault() {
	if cfg.MinLength <= 0 {
		cfg.MinLength = 8
	}

	if cfg.MaxLength <= 0 {
		cfg.MaxLength = 16
	}
}
