package passwordimpl

type Config struct {
	MinLength                  int `json:"min_length"`
	MaxLength                  int `json:"max_length"`
	MaxNumOfConsecutiveChars   int `json:"max_num_of_consecutive_chars"`
	MinNumOfKindOfPasswordChar int `json:"min_num_of_kind_of_password_char"`
}

func (cfg *Config) SetDefault() {
	if cfg.MinLength <= 0 {
		cfg.MinLength = 8
	}

	if cfg.MaxLength <= 0 {
		cfg.MaxLength = 20
	}

	if cfg.MaxNumOfConsecutiveChars <= 0 {
		cfg.MaxNumOfConsecutiveChars = 2
	}

	if cfg.MinNumOfKindOfPasswordChar <= 0 {
		cfg.MinNumOfKindOfPasswordChar = 3
	}
}
