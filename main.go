package main

import (
	"context"
	"os"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
	"github.com/opensourceways/app-cla-server/config"
	"github.com/opensourceways/app-cla-server/controllers"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/mongodb"
	"github.com/opensourceways/app-cla-server/pdf"
	"github.com/opensourceways/app-cla-server/robot/github"
	_ "github.com/opensourceways/app-cla-server/routers"
	"github.com/opensourceways/app-cla-server/util"
	"github.com/opensourceways/app-cla-server/worker"
	"github.com/opensourceways/robot-gitee-plugin-lib/interrupts"
)

const (
	serviceSign  = "sign"
	serviceRobot = "robot"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	enabledService := beego.AppConfig.String("enableservice")

	if enabledService != serviceSign && enabledService != serviceRobot {
		logs.Error("invaliid enableservice")
		os.Exit(1)
	}

	configFile := beego.AppConfig.String("appconf")

	if enabledService == serviceSign {
		startSignSerivce(configFile)
	} else {
		startRobotSerivce(configFile)
	}
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

	run(worker.GetEmailWorker().Shutdown)
}

func startMongoService(cfg *config.MongodbConfig) {
	c, err := mongodb.Initialize(cfg)
	if err != nil {
		logs.Error(err)
		os.Exit(1)
	}
	dbmodels.RegisterDB(c)
}

func startRobotSerivce(configPath string) {
	cfg, err := config.LoadRobotServiceeConfig(configPath)
	if err != nil {
		logs.Error(err)
		os.Exit(1)
	}

	startMongoService(&cfg.Mongodb)

	if err := github.InitGithubRobot(cfg.CLAPlatformURL, cfg.PlatformRobotConfigs); err != nil {
		logs.Error(err)
		os.Exit(1)
	}

	run(github.Stop)
}

func run(clear func()) {
	defer interrupts.WaitForGracefulShutdown()

	interrupts.OnInterrupt(func() {
		shutdown()
		if clear != nil {
			clear()
		}
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
