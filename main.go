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
	"github.com/opensourceways/app-cla-server/obs"
	_ "github.com/opensourceways/app-cla-server/obs/huaweicloud"
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

	if err := config.InitAppConfig(beego.AppConfig.String("app_conf")); err != nil {
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

	mongoClient, err := mongodb.Initialize(&AppConfig.Mongodb)
	if err != nil {
		beego.Error(err)
		os.Exit(1)
	}

	obsClient, err := obs.Initialize(AppConfig.OBS)
	if err != nil {
		beego.Error(err)
		os.Exit(1)
	}

	dbmodels.RegisterDB(struct {
		dbmodels.IModel
		dbmodels.IFile
	}{
		IModel: mongoClient,
		IFile:  obs.FileStorage{OBS: obsClient},
	})

	if err = email.Initialize(AppConfig.EmailPlatformConfigFile); err != nil {
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

	beego.Run()
}
