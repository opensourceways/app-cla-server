package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/opensourceways/obs-sdk/go/log"
	obsmetrics "github.com/opensourceways/obs-sdk/go/metrics"
	"github.com/opensourceways/obs-sdk/go/middleware/beegomw"

	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
	commondb "github.com/opensourceways/app-cla-server/common/infrastructure/mongodb"
	"github.com/opensourceways/app-cla-server/common/infrastructure/redisdb"
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/interrupts"
	"github.com/opensourceways/app-cla-server/pdf"
	_ "github.com/opensourceways/app-cla-server/routers"
	"github.com/opensourceways/app-cla-server/signing/domain"
	"github.com/opensourceways/app-cla-server/signing/domain/dp"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/smtpimpl"
	"github.com/opensourceways/app-cla-server/signing/watch"
	"github.com/opensourceways/app-cla-server/util"
	"github.com/opensourceways/app-cla-server/worker"
)

const serviceName = "app-cla-server"

var obsMetrics *obsmetrics.Metrics

type options struct {
	configFile string
}

func (o *options) Validate() error {
	if o.configFile == "" {
		return errors.New("missing config file")
	}

	return nil
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	fs.StringVar(
		&o.configFile, "config-file", "", "config file path.",
	)

	if err := fs.Parse(args); err != nil {
		log.Error("failed to parse flags", "error", err)
	}

	return o
}

func main() {
	log.Init(log.Config{Service: serviceName})

	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	o := gatherOptions(
		flag.NewFlagSet(os.Args[0], flag.ExitOnError),
		os.Args[1:]...,
	)
	if err := o.Validate(); err != nil {
		log.Error("invalid options", "error", err)

		return
	}

	cfg := loadConfig(o.configFile)
	if cfg == nil {
		return
	}

	startSignSerivce(cfg)
}

func loadConfig(f string) *config.Config {
	cfg, err := config.Load(f)
	err1 := os.Remove(f)

	if err2 := util.MultiErrors(err, err1); err2 != nil {
		log.Error("failed to load config or remove temp file", "error", err2)

		return nil
	}

	return &cfg
}

func startSignSerivce(cfg *config.Config) {
	dp.Init(&cfg.Domain.DomainPrimitive)
	domain.Init(&cfg.Domain.Config)

	if err := emailtmpl.Init(); err != nil {
		log.Error("failed to init email templates", "error", err)
		return
	}

	smtpimpl.Init(&cfg.SMTP)

	if err := platformAuth.Initialize(&cfg.CodePlatform); err != nil {
		log.Error("failed to init code platform auth", "error", err)
		return
	}

	if err := pdf.InitPDFGenerator(&cfg.PDF); err != nil {
		log.Error("failed to init pdf generator", "error", err)
		return
	}

	if err := commondb.Init(&cfg.Mongodb.DB); err != nil {
		log.Error("failed to init mongodb", "error", err)
		return
	}

	defer exitMongoService()

	if err := redisdb.Init(&cfg.Redisdb.DB); err != nil {
		log.Error("failed to init redis", "error", err)
		return
	}

	defer exitRedisService()

	// must run after init mongodb
	if err := initSigning(cfg); err != nil {
		log.Error("failed to init signing", "error", err)

		return
	}

	// watch.Start was invoked in initSigning
	defer watch.Stop()
	defer watch.CLAUpdateWatchStop()
	defer watch.NotifyAdminWatchStop()

	worker.Init(pdf.GetPDFGenerator())
	defer worker.Exit()

	obsMetrics = obsmetrics.New(obsmetrics.Config{Service: serviceName})
	worker.InitBusinessMetrics(obsMetrics)

	beego.InsertFilterChain("/*", beegomw.Middleware(beegomw.Options{Metrics: obsMetrics}))
	beego.Handler("/metrics", obsMetrics.Handler())

	run()
}

func exitMongoService() {
	if err := commondb.Close(); err != nil {
		log.Error("failed to close mongodb", "error", err)
	}
}

func exitRedisService() {
	if err := redisdb.Close(); err != nil {
		log.Error("failed to close redis", "error", err)
	}
}

func run() {
	defer interrupts.WaitForGracefulShutdown()

	interrupts.OnInterrupt(func() {
		shutdown()
	})

	beego.Run()
}

func shutdown() {
	log.Info("server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := beego.BeeApp.Server.Shutdown(ctx); err != nil {
		log.Error("failed to shut down server", "error", err)
	}
	cancel()
}
