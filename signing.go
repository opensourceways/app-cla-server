package main

import (
	"github.com/opensourceways/app-cla-server/common/infrastructure/mongodb"
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/adapter"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/accesstokenservice"
	"github.com/opensourceways/app-cla-server/signing/domain/claservice"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/domain/emailcredential"
	"github.com/opensourceways/app-cla-server/signing/domain/userservice"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/encryptionimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/gmailimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/localclaimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/passwordimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/randombytesimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/randomcodeimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/repositoryimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/symmetricencryptionimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/txmailimpl"
)

func initSigning() error {
	cfg := &config.AppConfig.SigningConfig

	symmetric, err := symmetricencryptionimpl.NewSymmetricEncryptionImpl(&cfg.Symmetric)
	if err != nil {
		return err
	}

	dp.Init(&cfg.Domain.DomainPrimitive)

	repo := repositoryimpl.NewCorpSigning(
		mongodb.DAO(cfg.Mongodb.Collections.CorpSigning),
	)

	userService := userservice.NewUserService(
		repositoryimpl.NewUser(
			mongodb.DAO(cfg.Mongodb.Collections.CorpSigning),
		),
		encryptionimpl.NewEncryptionImpl(),
		passwordimpl.NewPasswordImpl(&cfg.Password),
	)

	models.RegisterCorpAdminAdatper(
		adapter.NewCorpAdminAdapter(app.NewCorpAdminService(repo, userService)),
	)

	models.RegisterCorpSigningAdapter(
		adapter.NewCorpSigningAdapter(app.NewCorpSigningService(repo)),
	)

	models.RegisterEmployeeSigningAdapter(
		adapter.NewEmployeeSigningAdapter(app.NewEmployeeSigningService(repo)),
	)

	models.RegisterEmployeeManagerAdapter(
		adapter.NewEmployeeManagerAdapter(app.NewEmployeeManagerService(repo, userService)),
	)

	models.RegisterCorpEmailDomainAdapter(
		adapter.NewCorpEmailDomainAdapter(app.NewCorpEmailDomainService(repo)),
	)

	models.RegisterCorpPDFAdapter(
		adapter.NewCorpPDFAdapter(app.NewCorpPDFService(repo)),
	)

	vcService := vcservice.NewVCService(
		repositoryimpl.NewVerificationCode(
			mongodb.DAO(cfg.Mongodb.Collections.VerificationCode),
		),
		randomcodeimpl.NewRandomCodeImpl(),
	)

	models.RegisterVerificationCodeAdapter(
		adapter.NewVerificationCodeAdapter(app.NewVerificationCodeService(
			vcService,
		)),
	)

	models.RegisterUserAdapter(
		adapter.NewUserAdapter(app.NewUserService(userService, repo, symmetric, vcService)),
	)

	models.RegisterIndividualSigningAdapter(
		adapter.NewIndividualSigningAdapter(app.NewIndividualSigningService(
			repositoryimpl.NewIndividualSigning(
				mongodb.DAO(cfg.Mongodb.Collections.IndividualSigning),
			),
			repo,
		)),
	)

	// email credential
	ecRepo := repositoryimpl.NewEmailCredential(
		mongodb.DAO(cfg.Mongodb.Collections.EmailCredential),
	)

	echelper := emailcredential.NewEmailCredential(ecRepo, symmetric)

	gmailimpl.RegisterEmailService(echelper.Find)
	txmailimpl.RegisterEmailService(echelper.Find)

	models.RegisterEmailCredentialAdapter(
		adapter.NewEmailCredentialAdapter(
			app.NewEmailCredentialService(echelper), ecRepo,
		),
	)

	// access token
	at := accesstokenservice.NewAccessTokenService(
		nil,
		config.AppConfig.APITokenExpiry,
		encryptionimpl.NewEncryptionImpl(),
		randombytesimpl.NewRandomBytesImpl(),
	)

	models.RegisterAccessTokenAdapter(
		adapter.NewAccessTokenAdapter(app.NewAccessTokenService(at)),
	)

	// link
	linkRepo := repositoryimpl.NewLink(
		mongodb.DAO(cfg.Mongodb.Collections.Link),
		mongodb.DAO(cfg.Mongodb.Collections.CLA),
	)
	cla := claservice.NewCLAService(linkRepo, localclaimpl.NewLocalCLAImpl(&cfg.LocalCLA))

	claAapter := adapter.NewCLAAdapter(app.NewCLAService(linkRepo, cla), 0) // TODO

	models.RegisterCLAAdapter(claAapter)

	models.RegisterLinkAdapter(
		adapter.NewLinkAdapter(
			app.NewLinkService(linkRepo, cla, echelper),
			claAapter,
		),
	)

	return nil
}
