package main

import (
	"os"

	"github.com/astaxie/beego"

	platformAuth "github.com/opensourceways/app-cla-server/code-platform-auth"
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

	c, err := mongodb.RegisterDatabase(
		beego.AppConfig.String("mongodb_conn"),
		beego.AppConfig.String("mongodb_db"))
	if err != nil {
		beego.Error(err)
		os.Exit(1)
	}
	dbmodels.RegisterDB(c)

	path := beego.AppConfig.String("email_platforms")
	if err = email.RegisterPlatform(path); err != nil {
		beego.Error(err)
		os.Exit(1)
	}

	path = beego.AppConfig.String("code_platforms")
	if err := platformAuth.RegisterPlatform(path); err != nil {
		beego.Error(err)
		os.Exit(1)
	}

	language := beego.AppConfig.String("blank_signature::language")
	path = beego.AppConfig.String("blank_signature::pdf")
	if err := pdf.UploadBlankSignature(language, path); err != nil {
		beego.Error(err)
		os.Exit(1)
	}

	if err := pdf.InitPDFGenerator(
		beego.AppConfig.String("python_bin"),
		beego.AppConfig.String("pdf_out_dir"),
		beego.AppConfig.String("pdf_org_signature_dir"),
		beego.AppConfig.String("pdf_template_corporation::welcome"),
		beego.AppConfig.String("pdf_template_corporation::declaration"),
	); err != nil {
		beego.Error(err)
		os.Exit(1)
	}

	worker.InitEmailWorker(pdf.GetPDFGenerator())

	beego.Run()
}
