package dp

import "strings"

var config Config

func Init(cfg *Config) {
	config = *cfg
}

type Config struct {
	MaxLengthOfName     int `json:"max_length_of_name"        required:"true"`
	MaxLengthOfTitle    int `json:"max_length_of_title"       required:"true"`
	MaxLengthOfEmail    int `json:"max_length_of_email"       required:"true"`
	MaxLengthOfAccount  int `json:"max_length_of_account"     required:"true"`
	MaxLengthOfCorpName int `json:"max_length_of_corp_name"   required:"true"`

	SupportedLanguages           []string `json:"supported_languages"`
	SupportedCorpCLAFields       []string `json:"supported_corp_cla_fields"`
	SupportedIndividualCLAFields []string `json:"supported_individual_cla_fields"`

	supportedLanguages map[string]string
}

func (cfg *Config) SetDefault() {
	if len(cfg.SupportedLanguages) == 0 {
		cfg.SupportedLanguages = []string{"Chinese", "English"}
	}

	if len(cfg.SupportedCorpCLAFields) == 0 {
		cfg.SupportedCorpCLAFields = []string{
			"fax",
			"title",
			"email",
			"address",
			"telephone",
			"authorized",
			"corporationName",
		}
	}

	if len(cfg.SupportedIndividualCLAFields) == 0 {
		cfg.SupportedIndividualCLAFields = []string{
			"name", "email",
		}
	}
}

func (cfg *Config) Validate() error {
	v := map[string]string{}
	for _, item := range cfg.SupportedLanguages {
		v[strings.ToLower(item)] = item
	}

	cfg.supportedLanguages = v

	return nil
}

func (cfg *Config) getLanguage(v string) string {
	return cfg.supportedLanguages[strings.ToLower(v)]
}

func (cfg *Config) isValidCorpCLAField(v string) bool {
	return cfg.has(v, cfg.SupportedCorpCLAFields)
}

func (cfg *Config) isValidIndividualCLAField(v string) bool {
	return cfg.has(v, cfg.SupportedIndividualCLAFields)
}

func (cfg *Config) has(v string, items []string) bool {
	for _, s := range items {
		if v == s {
			return true
		}
	}

	return false
}
