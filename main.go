package main

import (
	"github.com/astaxie/beego"

	platformAuth "github.com/zengchen1024/cla-server/code-platform-auth"
	"github.com/zengchen1024/cla-server/dbmodels"
	"github.com/zengchen1024/cla-server/email"
	"github.com/zengchen1024/cla-server/models"
	"github.com/zengchen1024/cla-server/mongodb"
	"github.com/zengchen1024/cla-server/pdf"
	_ "github.com/zengchen1024/cla-server/routers"
	"github.com/zengchen1024/cla-server/worker"
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
		return
	}

	models.RegisterDB(c)
	dbmodels.RegisterDB(c)

	path := beego.AppConfig.String("gmail::credentials")
	webRedirectDir := beego.AppConfig.String("gmail::web_redirect_dir")
	if err = email.RegisterPlatform("gmail", path, webRedirectDir); err != nil {
		beego.Info(err)
		return
	}

	path = beego.AppConfig.String("gitee::credentials")
	if err := platformAuth.RegisterPlatform("gitee", path); err != nil {
		beego.Info(err)
		return
	}

	language := beego.AppConfig.String("blank_signature::language")
	path = beego.AppConfig.String("blank_signature::pdf")
	if err := pdf.UploadBlankSignature(language, path); err != nil {
		beego.Info(err)
		return
	}

	if err := pdf.InitPDFGenerator(
		beego.AppConfig.String("python_bin"),
		beego.AppConfig.String("pdf_out_dir"),
		beego.AppConfig.String("pdf_org_signature_dir"),
		beego.AppConfig.String("pdf_template_corporation::welcome"),
		beego.AppConfig.String("pdf_template_corporation::declaration"),
	); err != nil {
		beego.Info(err)
		return
	}

	worker.InitEmailWorker(pdf.GetPDFGenerator())

	beego.Run()
}
