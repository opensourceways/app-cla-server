package main

import (
	"github.com/opensourceways/app-cla-server/common/infrastructure/mongodb"
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/adapter"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/repositoryimpl"
)

func initSigning() {
	cfg := &config.AppConfig.SigningConfig

	dp.Init(&cfg.Domain.DomainPrimitive)

	cs := adapter.NewCorpSigningAdapter(app.NewCorpSigningService(
		repositoryimpl.NewCorpSigning(
			mongodb.DAO(cfg.Mongodb.Collections.CorpSigning),
		),
	))

	models.Init(cs)
}
