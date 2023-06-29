package main

import (
	"github.com/opensourceways/app-cla-server/common/infrastructure/mongodb"
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/adapter"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/userservice"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/encryptionimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/passwordimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/randomcodeimpl"
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

	em := adapter.NewEmployeeManagerAdapter(app.NewEmployeeManagerService(repo, userService))

	ed := adapter.NewCorpEmailDomainAdapter(app.NewCorpEmailDomainService(repo))

	cp := adapter.NewCorpPDFAdapter(app.NewCorpPDFService(repo))

	vc := adapter.NewVerificationCodeAdapter(app.NewVerificationCodeService(
		vcservice.NewVCService(
			repositoryimpl.NewVerificationCode(
				mongodb.DAO(cfg.Mongodb.Collections.VerificationCode),
			),
			randomcodeimpl.NewRandomCodeImpl(),
		),
	))

	models.Init(ua, cp, ca, cs, es, em, ed, vc)
}
