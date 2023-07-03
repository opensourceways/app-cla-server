package main

import (
	"github.com/opensourceways/app-cla-server/common/infrastructure/mongodb"
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/adapter"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/accesstokenservice"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/emailcredential"
	"github.com/opensourceways/app-cla-server/signing/domain/userservice"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/encryptionimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/gmailimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/passwordimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/randomcodeimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/randomstrimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/repositoryimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/symmetricencryptionimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/txmailimpl"
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

	ua := adapter.NewUserAdapter(app.NewUserService(userService, repo))

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

	is := adapter.NewIndividualSigningAdapter(app.NewIndividualSigningService(
		repositoryimpl.NewIndividualSigning(
			mongodb.DAO(cfg.Mongodb.Collections.IndividualSigning),
		),
		repo,
	))

	ecRepo := repositoryimpl.NewEmailCredential(
		mongodb.DAO(cfg.Mongodb.Collections.EmailCredential),
	)

	symmetricEncrypt := symmetricencryptionimpl.NewSymmetricEncryptionImpl()

	echelper := emailcredential.NewEmailCredential(ecRepo, symmetricEncrypt)

	gmailimpl.RegisterEmailService(echelper.Find)
	txmailimpl.RegisterEmailService(echelper.Find)

	ec := adapter.NewEmailCredentialAdapter(
		app.NewEmailCredentialService(echelper), ecRepo,
	)

	models.Init(ua, cp, ca, cs, ec, es, em, ed, vc, is)

	at := accesstokenservice.NewAccessTokenService(
		nil,
		config.AppConfig.APITokenExpiry,
		symmetricEncrypt, randomstrimpl.NewRandomStrImpl(),
	)

	models.RegisterAccessTokenAdapter(
		adapter.NewAccessTokenAdapter(app.NewAccessTokenService(at)),
	)
}
