package main

import (
	"github.com/opensourceways/app-cla-server/common/infrastructure/mongodb"
	"github.com/opensourceways/app-cla-server/common/infrastructure/redisdb"
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/controllers"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/signing/adapter"
	"github.com/opensourceways/app-cla-server/signing/app"
	"github.com/opensourceways/app-cla-server/signing/domain/accesstokenservice"
	"github.com/opensourceways/app-cla-server/signing/domain/claservice"
	"github.com/opensourceways/app-cla-server/signing/domain/emailcredential"
	"github.com/opensourceways/app-cla-server/signing/domain/loginservice"
	"github.com/opensourceways/app-cla-server/signing/domain/userservice"
	"github.com/opensourceways/app-cla-server/signing/domain/vcservice"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/accesstokenimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/encryptionimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/limiterimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/localclaimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/loginimpl"
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

	pi := passwordimpl.NewPasswordImpl(&cfg.Password)
	ur := repositoryimpl.NewUser(
		mongodb.DAO(cfg.Mongodb.Collections.User),
	)
	userService := userservice.NewUserService(
		ur,
		encryptionimpl.NewEncryptionImpl(),
		pi,
	)

	loginService := loginservice.NewLoginService(
		ur,
		loginimpl.NewLoginImpl(redisdb.DAO(), &cfg.Redisdb.Login),
		encryptionimpl.NewEncryptionImpl(),
		pi,
	)

	vcService := vcservice.NewVCService(
		repositoryimpl.NewVerificationCode(
			mongodb.DAO(cfg.Mongodb.Collections.VerificationCode),
		),
		limiterimpl.NewLimiterImpl(redisdb.DAO()),
		randomcodeimpl.NewRandomCodeImpl(),
	)

	models.RegisterCorpAdminAdatper(
		adapter.NewCorpAdminAdapter(app.NewCorpAdminService(repo, userService)),
	)

	interval := cfg.Domain.Config.GetIntervalOfCreatingVC()

	models.RegisterCorpSigningAdapter(
		adapter.NewCorpSigningAdapter(
			app.NewCorpSigningService(repo, vcService, interval),
			cfg.Domain.Config.InvalidCorpEmailDomains(),
		),
	)

	models.RegisterEmployeeSigningAdapter(
		adapter.NewEmployeeSigningAdapter(
			app.NewEmployeeSigningService(repo, vcService, interval),
		),
	)

	models.RegisterEmployeeManagerAdapter(
		adapter.NewEmployeeManagerAdapter(app.NewEmployeeManagerService(repo, userService)),
	)

	models.RegisterCorpEmailDomainAdapter(
		adapter.NewCorpEmailDomainAdapter(app.NewCorpEmailDomainService(vcService, repo)),
	)

	models.RegisterCorpPDFAdapter(
		adapter.NewCorpPDFAdapter(app.NewCorpPDFService(repo)),
	)

	models.RegisterUserAdapter(
		adapter.NewUserAdapter(
			app.NewUserService(userService, loginService, repo, symmetric, vcService, interval),
		),
	)

	individual := repositoryimpl.NewIndividualSigning(
		mongodb.DAO(cfg.Mongodb.Collections.IndividualSigning),
	)

	models.RegisterIndividualSigningAdapter(
		adapter.NewIndividualSigningAdapter(app.NewIndividualSigningService(
			vcService,
			individual,
			repo,
			interval,
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

	// access token
	at := accesstokenservice.NewAccessTokenService(
		accesstokenimpl.NewAccessTokenImpl(redisdb.DAO(), &cfg.Redisdb.AccessToken),
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
		app.NewCLAService(linkRepo, cla, repo, individual),
		cfg.Domain.MaxSizeOfCLAContent,
		cfg.Domain.FileTypeOfCLAContent,
		cfg.Domain.SourceOfCLAPDF,
	)

	models.RegisterCLAAdapter(claAapter)

	models.RegisterLinkAdapter(
		adapter.NewLinkAdapter(
			app.NewLinkService(linkRepo, cla, repo, individual, echelper),
			claAapter,
		),
	)

	//
	controllers.Init(
		&cfg.API,
		repositoryimpl.NewOrg(
			mongodb.DAO(cfg.Mongodb.Collections.Org),
		),
	)

	return nil
}
