package domain

var config Config

func Init(cfg *Config) {
	config = *cfg
}

type Config struct {
	MaxNumOfEmployeeManager int `json:"max_num_of_employee_manager"`
}

func (cfg *Config) SetDefault() {
	if cfg.MaxNumOfEmployeeManager <= 0 {
		cfg.MaxNumOfEmployeeManager = 5
	}
}
