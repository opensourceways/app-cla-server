package main

import (
	"github.com/opensourceways/app-cla-server/common/infrastructure/mongodb"
	"github.com/opensourceways/app-cla-server/common/infrastructure/redisdb"
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/adapter"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/accesstokenservice"
	"github.com/opensourceways/app-cla-server/signing/domain/claservice"
	"github.com/opensourceways/app-cla-server/signing/domain/emailcredential"
	"github.com/opensourceways/app-cla-server/signing/domain/userservice"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/accesstokenimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/encryptionimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/gmailimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/localclaimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/passwordimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/randombytesimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/randomcodeimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/repositoryimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/smtpimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/symmetricencryptionimpl"
)

func initSigning(cfg *config.Config) error {
	symmetric, err := symmetricencryptionimpl.NewSymmetricEncryptionImpl(&cfg.Symmetric)
	if err != nil {
		return err
	}

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
		adapter.NewCorpSigningAdapter(
			app.NewCorpSigningService(repo),
			cfg.Domain.Config.InvalidCorpEmailDomains(),
		),
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

	// smtp
	smtpimpl.RegisterEmailService(echelper.Find)

	models.RegisterSMTPAdapter(
		adapter.NewSMTPAdapter(app.NewSMTPService(vcService, echelper)),
	)

	// gmail
	gmailimpl.RegisterEmailService(echelper.Find)

	models.RegisterGmailAdapter(
		adapter.NewGmailAdapter(
			app.NewGmailService(echelper, ecRepo),
		),
	)

	// access token
	at := accesstokenservice.NewAccessTokenService(
		accesstokenimpl.NewAccessTokenImpl(redisdb.DAO(), &cfg.Redisdb.Config),
		cfg.Domain.Config.AccessTokenExpiry,
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

	claAapter := adapter.NewCLAAdapter(
		app.NewCLAService(linkRepo, cla),
		cfg.Domain.MaxSizeOfCLAContent,
		cfg.Domain.FileTypeOfCLAContent,
	)

	models.RegisterCLAAdapter(claAapter)

	models.RegisterLinkAdapter(
		adapter.NewLinkAdapter(
			app.NewLinkService(linkRepo, cla, echelper),
			claAapter,
		),
	)

	return nil
}
