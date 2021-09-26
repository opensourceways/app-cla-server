package main

import (
	"os"

	"github.com/astaxie/beego"

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

	enabledService := beego.AppConfig.String("EnableService")

	if enabledService != serviceSign && enabledService != serviceRobot {
		beego.Error("invaliid EnableService")
		os.Exit(1)
	}

	configFile := beego.AppConfig.String("app_conf")

	if enabledService == serviceSign {
		startSignSerivce(configFile)
	} else {
		startRobotSerivce(configFile)
	}

	beego.Run()
}

func startSignSerivce(configPath string) {
	if err := config.InitAppConfig(configPath); err != nil {
		beego.Error(err)
		os.Exit(1)
	}
	AppConfig := config.AppConfig

	path := util.GenFilePath(AppConfig.PDFOutDir, "tmp")
	if util.IsNotDir(path) {
		err := os.Mkdir(path, 0732)
		if err != nil {
			beego.Error(err)
			os.Exit(1)
		}
	}

	startMongoService(&AppConfig.Mongodb)

	if err := email.Initialize(AppConfig.EmailPlatformConfigFile); err != nil {
		beego.Error(err)
		os.Exit(1)
	}

	if err := platformAuth.Initialize(AppConfig.CodePlatformConfigFile); err != nil {
		beego.Error(err)
		os.Exit(1)
	}

	if err := pdf.InitPDFGenerator(
		AppConfig.PythonBin,
		AppConfig.PDFOutDir,
		AppConfig.PDFOrgSignatureDir,
	); err != nil {
		beego.Error(err)
		os.Exit(1)
	}

	worker.InitEmailWorker(pdf.GetPDFGenerator())

	if err := controllers.LoadLinks(); err != nil {
		beego.Error(err)
		os.Exit(1)
	}
}

func startMongoService(cfg *config.MongodbConfig) {
	c, err := mongodb.Initialize(cfg)
	if err != nil {
		beego.Error(err)
		os.Exit(1)
	}
	dbmodels.RegisterDB(c)
}

func startRobotSerivce(configPath string) {
	cfg, err := config.LoadRobotServiceeConfig(configPath)
	if err != nil {
		beego.Error(err)
		os.Exit(1)
	}

	if err := github.InitGithubRobot(cfg.CLAPlatformURL, cfg.PlatformRobotConfigs); err != nil {
		beego.Error(err)
		os.Exit(1)
	}

	startMongoService(&cfg.Mongodb)
}
