package config

import (
	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
	"github.com/opensourceways/app-cla-server/common/infrastructure/mongodb"
	"github.com/opensourceways/app-cla-server/controllers"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/gmailimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/localclaimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/passwordimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/repositoryimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/symmetricencryptionimpl"
	"github.com/opensourceways/app-cla-server/util"
)

func Load(path string) (cfg Config, err error) {
	if err = util.LoadFromYaml(path, &cfg); err != nil {
		return
	}

	cfg.setDefault()

	if err = cfg.validate(); err != nil {
		return
	}

	return
}

type configValidate interface {
	Validate() error
}

type configSetDefault interface {
	SetDefault()
}

type domainConfig struct {
	domain.Config

	DomainPrimitive dp.Config `json:"domain_primitive"  required:"true"`
}

type mongodbConfig struct {
	DB mongodb.Config `json:"db" required:"true"`

	repositoryimpl.Config
}

type Config struct {
	PDF          pdf.Config                     `json:"pdf"             required:"true"`
	API          controllers.Config             `json:"api"             required:"true"`
	Gmail        gmailimpl.Config               `json:"gmail"           required:"true"`
	Domain       domainConfig                   `json:"domain"          required:"true"`
	Mongodb      mongodbConfig                  `json:"mongodb"         required:"true"`
	Password     passwordimpl.Config            `json:"password"        required:"true"`
	LocalCLA     localclaimpl.Config            `json:"local_cla"       required:"true"`
	Symmetric    symmetricencryptionimpl.Config `json:"symmetric"       required:"true"`
	CodePlatform platformAuth.Config            `json:"code_platform"   required:"true"`
}

func (cfg *Config) configItems() []interface{} {
	return []interface{}{
		&cfg.PDF,
		&cfg.API,
		&cfg.Gmail,
		&cfg.Mongodb.DB,
		&cfg.Mongodb.Config,
		&cfg.Password,
		&cfg.LocalCLA,
		&cfg.Symmetric,
		&cfg.Domain.Config,
		&cfg.Domain.DomainPrimitive,
		&cfg.CodePlatform,
	}
}

func (cfg *Config) setDefault() {
	items := cfg.configItems()
	for _, i := range items {
		if f, ok := i.(configSetDefault); ok {
			f.SetDefault()
		}
	}
}

func (cfg *Config) validate() error {
	items := cfg.configItems()
	for _, i := range items {
		if f, ok := i.(configValidate); ok {
			if err := f.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}
