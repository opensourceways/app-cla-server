package main

import (
	"os"

	"github.com/astaxie/beego"

	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/mongodb"
	"github.com/opensourceways/app-cla-server/pdf"
	_ "github.com/opensourceways/app-cla-server/routers"
	"github.com/opensourceways/app-cla-server/worker"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	if err := conf.InitAppConfig(); err != nil {
		beego.Error(err)
		os.Exit(1)
	}
	AppConfig := conf.AppConfig

	c, err := mongodb.RegisterDatabase(AppConfig.MongodbConn, AppConfig.DBName)
	if err != nil {
		beego.Error(err)
		os.Exit(1)
	}
	dbmodels.RegisterDB(c)

	if err = email.RegisterPlatform(AppConfig.EmailPlatformConfigFile); err != nil {
		beego.Error(err)
		os.Exit(1)
	}

	if err := platformAuth.RegisterPlatform(AppConfig.CodePlatformConfigFile); err != nil {
		beego.Error(err)
		os.Exit(1)
	}

	if err := pdf.GenBlankSignaturePage(); err != nil {
		beego.Info(err)
		return
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

	beego.Run()
}
