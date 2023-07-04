package main

import (
	"context"
	"os"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/opensourceways/server-common-lib/interrupts"

	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
	commondb "github.com/opensourceways/app-cla-server/common/infrastructure/mongodb"
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/controllers"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/mongodb"
	"github.com/opensourceways/app-cla-server/pdf"
	_ "github.com/opensourceways/app-cla-server/routers"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/emailtmpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/gmailimpl"
	"github.com/opensourceways/app-cla-server/signing/infrastructure/txmailimpl"
	"github.com/opensourceways/app-cla-server/util"
	"github.com/opensourceways/app-cla-server/worker"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	configFile, err := beego.AppConfig.String("appconf")
	if err != nil {
		logs.Error(err)
		return
	}

	startSignSerivce(configFile)
}

func startSignSerivce(configPath string) {
	if err := config.InitAppConfig(configPath); err != nil {
		logs.Error(err)
		return
	}
	cfg := config.AppConfig

	path := util.GenFilePath(cfg.PDFOutDir, "tmp")
	if util.IsNotDir(path) {
		if err := os.Mkdir(path, 0732); err != nil {
			logs.Error(err)
			return
		}
	}

	if err := emailtmpl.Init(); err != nil {
		logs.Error(err)
		return
	}

	if err := gmailimpl.Init(&cfg.SigningConfig.Gmail); err != nil {
		logs.Error(err)
		return
	}

	txmailimpl.Init()

	if err := platformAuth.Initialize(cfg.CodePlatformConfigFile); err != nil {
		logs.Error(err)
		return
	}

	err := pdf.InitPDFGenerator(cfg.PythonBin, cfg.PDFOutDir, cfg.PDFOrgSignatureDir)
	if err != nil {
		logs.Error(err)
		return
	}

	if err := controllers.LoadLinks(); err != nil {
		logs.Error(err)
		return
	}

	if err := controllers.Init(); err != nil {
		logs.Error(err)
		return
	}

	if err := startMongoService(); err != nil {
		logs.Error(err)
		return
	}

	defer exitMongoService()

	// must run after init mongodb
	if err := initSigning(); err != nil {
		logs.Error(err)

		return
	}

	worker.Init(pdf.GetPDFGenerator())
	defer worker.Exit()

	run()
}

func startMongoService() error {
	cfg := &config.AppConfig.SigningConfig

	if err := commondb.Init(&cfg.Mongodb.DB); err != nil {
		return err
	}

	c := mongodb.Initialize(commondb.Collection(), &config.AppConfig.Mongodb)
	dbmodels.RegisterDB(c)

	return nil
}

func exitMongoService() {
	if err := commondb.Close(); err != nil {
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
