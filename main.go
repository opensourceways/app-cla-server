package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"

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
	"github.com/opensourceways/app-cla-server/util"
	"github.com/opensourceways/app-cla-server/worker"
)

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
		logs.Error(err)
	}

	return o
}

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	o := gatherOptions(
		flag.NewFlagSet(os.Args[0], flag.ExitOnError),
		os.Args[1:]...,
	)
	if err := o.Validate(); err != nil {
		logs.Error("Invalid options, err:%s", err.Error())

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
		logs.Error(err2)

		return nil
	}

	return &cfg
}

func startSignSerivce(cfg *config.Config) {
	dp.Init(&cfg.Domain.DomainPrimitive)
	domain.Init(&cfg.Domain.Config)

	if err := emailtmpl.Init(); err != nil {
		logs.Error(err)
		return
	}

	smtpimpl.Init(&cfg.SMTP)

	if err := platformAuth.Initialize(&cfg.CodePlatform); err != nil {
		logs.Error(err)
		return
	}

	if err := pdf.InitPDFGenerator(&cfg.PDF); err != nil {
		logs.Error(err)
		return
	}

	if err := commondb.Init(&cfg.Mongodb.DB); err != nil {
		logs.Error(err)
		return
	}

	defer exitMongoService()

	if err := redisdb.Init(&cfg.Redisdb.DB); err != nil {
		logs.Error(err)
		return
	}

	defer exitRedisService()

	// must run after init mongodb
	if err := initSigning(cfg); err != nil {
		logs.Error(err)

		return
	}

	worker.Init(pdf.GetPDFGenerator())
	defer worker.Exit()

	run()
}

func exitMongoService() {
	if err := commondb.Close(); err != nil {
		logs.Error(err)
	}
}

func exitRedisService() {
	if err := redisdb.Close(); err != nil {
		logs.Error(err)
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
	logs.Info("server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := beego.BeeApp.Server.Shutdown(ctx); err != nil {
		logs.Error("error to shut down server, err:", err.Error())
	}
	cancel()
}
