package config

import (
	"github.com/opensourceways/app-cla-server/common/infrastructure/mongodb"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/gmailimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/repositoryimpl"
)

type configValidate interface {
	Validate() error
}

type configSetDefault interface {
	SetDefault()
}

type domainConfig struct {
	DomainPrimitive dp.Config `json:"domain_primitive"  required:"true"`
}

type mongodbConfig struct {
	DB mongodb.Config `json:"db" required:"true"`

	repositoryimpl.Config
}

type signingConfig struct {
	Gmail   gmailimpl.Config `json:"gmail"       required:"true"`
	Domain  domainConfig     `json:"domain"      required:"true"`
	Mongodb mongodbConfig    `json:"mongodb"     required:"true"`
}

func (cfg *signingConfig) configItems() []interface{} {
	return []interface{}{
		&cfg.Gmail,
		&cfg.Mongodb.DB,
		&cfg.Mongodb.Config,
		&cfg.Domain.DomainPrimitive,
	}
}

func (cfg *signingConfig) setDefault() {
	items := cfg.configItems()
	for _, i := range items {
		if f, ok := i.(configSetDefault); ok {
			f.SetDefault()
		}
	}
}

func (cfg *signingConfig) validate() error {
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
