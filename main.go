package main

import (
	"context"
	"os"
	"time"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/adapter/logs"
	"github.com/opensourceways/server-common-lib/interrupts"

	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/controllers"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/mongodb"
	"github.com/opensourceways/app-cla-server/pdf"
	_ "github.com/opensourceways/app-cla-server/routers"
	"github.com/opensourceways/app-cla-server/util"
	"github.com/opensourceways/app-cla-server/worker"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	configFile := beego.AppConfig.String("appconf")

	startSignSerivce(configFile)
}

func startSignSerivce(configPath string) {
	if err := config.InitAppConfig(configPath); err != nil {
		logs.Error(err)
		os.Exit(1)
	}
	cfg := config.AppConfig

	path := util.GenFilePath(cfg.PDFOutDir, "tmp")
	if util.IsNotDir(path) {
		if err := os.Mkdir(path, 0732); err != nil {
			logs.Error(err)
			os.Exit(1)
		}
	}

	startMongoService(&cfg.Mongodb)
	defer exitMongoService()

	if err := email.Initialize(cfg.EmailPlatformConfigFile); err != nil {
		logs.Error(err)
		os.Exit(1)
	}

	if err := platformAuth.Initialize(cfg.CodePlatformConfigFile); err != nil {
		logs.Error(err)
		os.Exit(1)
	}

	err := pdf.InitPDFGenerator(cfg.PythonBin, cfg.PDFOutDir, cfg.PDFOrgSignatureDir)
	if err != nil {
		logs.Error(err)
		os.Exit(1)
	}

	if err := controllers.LoadLinks(); err != nil {
		logs.Error(err)
		os.Exit(1)
	}

	worker.InitEmailWorker(pdf.GetPDFGenerator())
	defer worker.GetEmailWorker().Shutdown()

	run()
}

func startMongoService(cfg *config.MongodbConfig) {
	c, err := mongodb.Initialize(cfg)
	if err != nil {
		logs.Error(err)
		os.Exit(1)
	}
	dbmodels.RegisterDB(c)
}

func exitMongoService() {
	err := dbmodels.GetDB().Close()
	logs.Info("mongo exit, err:%v", err)
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
