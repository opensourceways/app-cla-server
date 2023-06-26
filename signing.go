package main

import (
	"github.com/opensourceways/app-cla-server/common/infrastructure/mongodb"
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/adapter"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/userservice"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/encryptionimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/passwordimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/repositoryimpl"
)

func initSigning() {
	cfg := &config.AppConfig.SigningConfig

	dp.Init(&cfg.Domain.DomainPrimitive)

	repo := repositoryimpl.NewCorpSigning(
		mongodb.DAO(cfg.Mongodb.Collections.CorpSigning),
	)

	userService := userservice.NewUserService(
		repositoryimpl.NewUser(
			mongodb.DAO(cfg.Mongodb.Collections.CorpSigning),
		),
		encryptionimpl.NewEncryptionImpl(),
		passwordimpl.NewPasswordImpl(),
	)

	ua := adapter.NewUserAdapter(app.NewUserService(userService))

	ca := adapter.NewCorpAdminAdapter(app.NewCorpAdminService(repo, userService))

	cs := adapter.NewCorpSigningAdapter(app.NewCorpSigningService(repo))

	es := adapter.NewEmployeeSigningAdapter(app.NewEmployeeSigningService(repo))

	models.Init(ua, ca, cs, es)
}
