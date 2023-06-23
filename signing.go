package main

import (
	"github.com/beego/beego/v2/core/logs"

	"github.com/opensourceways/app-cla-server/common/infrastructure/mongodb"
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/adapter"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/repositoryimpl"
)

func initSigning() error {
	cfg := &config.AppConfig.SigningConfig

	if err := mongodb.Init(&cfg.Mongodb.DB); err != nil {
		return err
	}
	defer func() {
		if err := mongodb.Close(); err != nil {
			logs.Error(err)
		}
	}()

	dp.Init(&cfg.Domain.DomainPrimitive)

	cs := adapter.NewCorpSigningAdapter(app.NewCorpSigningService(
		repositoryimpl.NewCorpSigning(
			mongodb.DAO(cfg.Mongodb.Collections.CorpSigning),
		),
	))

	models.Init(cs)

	return nil
}

func existSigning() {
	if err := mongodb.Close(); err != nil {
		logs.Error(err)
	}
}
