package dp

import (
	"strings"

	"github.com/opensourceways/app-cla-server/util"
)

var config Config

func Init(cfg *Config) {
	config = *cfg
}

func GetCLAFileds(t CLAType, lang Language) []CLAField {
	return config.claFileds(t, lang)
}

type Config struct {
	MaxLengthOfName     int `json:"max_length_of_name"        required:"true"`
	MaxLengthOfTitle    int `json:"max_length_of_title"       required:"true"`
	MaxLengthOfEmail    int `json:"max_length_of_email"       required:"true"`
	MaxLengthOfAccount  int `json:"max_length_of_account"     required:"true"`
	MaxLengthOfCorpName int `json:"max_length_of_corp_name"   required:"true"`

	CLA                []claConfig `json:"cla" required:"true"`
	supportedLanguages map[string]string
}

func (cfg *Config) Validate() error {
	v := map[string]string{}

	for i := range cfg.CLA {
		cla := &cfg.CLA[i]

		if err := cla.validate(); err != nil {
			return err
		}

		items := cla.languages()
		for _, item := range items {
			v[strings.ToLower(item)] = item
		}
	}

	cfg.supportedLanguages = v

	return nil
}

func (cfg *Config) getLanguage(v string) string {
	return cfg.supportedLanguages[strings.ToLower(v)]
}

func (cfg *Config) claFileds(t CLAType, lang Language) []CLAField {
	for i := range cfg.CLA {
		if item := &cfg.CLA[i]; t.CLAType() == item.Type {
			return item.fields(lang)
		}
	}

	return nil
}

// claConfig
type claConfig struct {
	Type   string      `json:"type"     required:"true"`
	Fileds []claFields `json:"fields"   required:"true"`
}

func (cfg *claConfig) validate() error {
	_, err := NewCLAType(cfg.Type)

	return err
}

func (cfg *claConfig) languages() []string {
	v := make([]string, len(cfg.Fileds))

	for i := range cfg.Fileds {
		v[i] = cfg.Fileds[i].Language
	}

	return v
}

func (cfg *claConfig) fields(lang Language) []CLAField {
	for i := range cfg.Fileds {
		if item := &cfg.Fileds[i]; lang.Language() == item.Language {
			return item.Fileds
		}
	}

	return nil
}

// claFields
type claFields struct {
	Fileds   []CLAField `json:"fields"   required:"true"`
	Language string     `json:"language" required:"true"`
}

// CLAField
type CLAField struct {
	Type      string `json:"type"`
	Desc      string `json:"desc"`
	Title     string `json:"title"`
	MaxLength int    `json:"max_length"`
}

func (f *CLAField) IsValidValue(v string) bool {
	if util.StrLen(v) > f.MaxLength {
		return false
	}

	// TODO there is a case that the field can include string of XSS
	return !util.HasXSS(v)
}
