package config

import (
	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
	"github.com/opensourceways/app-cla-server/common/infrastructure/mongodb"
	"github.com/opensourceways/app-cla-server/common/infrastructure/redisdb"
	"github.com/opensourceways/app-cla-server/controllers"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/accesstokenimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/localclaimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/loginimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/passwordimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/repositoryimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/smtpimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/symmetricencryptionimpl"
	"github.com/opensourceways/app-cla-server/signing/watch"
	"github.com/opensourceways/app-cla-server/util"
)

func Load(path string) (cfg Config, err error) {
	if err = util.LoadFromYaml(path, &cfg); err != nil {
		return
	}

	cfg.setDefault()

	err = cfg.validate()

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

type redisdbConfig struct {
	DB          redisdb.Config         `json:"db"`
	Login       loginimpl.Config       `json:"login"`
	AccessToken accesstokenimpl.Config `json:"access_token"`
}

type Config struct {
	PDF          pdf.Config                     `json:"pdf"             required:"true"`
	API          controllers.Config             `json:"api"             required:"true"`
	SMTP         smtpimpl.Config                `json:"smtp"`
	Watch        watch.Config                   `json:"watch"`
	Domain       domainConfig                   `json:"domain"          required:"true"`
	Mongodb      mongodbConfig                  `json:"mongodb"         required:"true"`
	Redisdb      redisdbConfig                  `json:"redisdb"         required:"true"`
	Password     passwordimpl.Config            `json:"password"        required:"true"`
	LocalCLA     localclaimpl.Config            `json:"local_cla"       required:"true"`
	Symmetric    symmetricencryptionimpl.Config `json:"symmetric"       required:"true"`
	CodePlatform platformAuth.Config            `json:"code_platform"   required:"true"`
}

func (cfg *Config) configItems() []interface{} {
	return []interface{}{
		&cfg.PDF,
		&cfg.API,
		&cfg.SMTP,
		&cfg.Watch,
		&cfg.Domain.Config,
		&cfg.Domain.DomainPrimitive,
		&cfg.Mongodb.DB,
		&cfg.Mongodb.Config,
		&cfg.Redisdb.DB,
		&cfg.Redisdb.Login,
		&cfg.Redisdb.AccessToken,
		&cfg.Password,
		&cfg.LocalCLA,
		&cfg.Symmetric,
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
