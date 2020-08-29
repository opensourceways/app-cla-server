package main

import (
	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/dbmodels"
	"github.com/zengchen1024/cla-server/email"
	"github.com/zengchen1024/cla-server/models"
	"github.com/zengchen1024/cla-server/mongodb"
	"github.com/zengchen1024/cla-server/oauth2"
	_ "github.com/zengchen1024/cla-server/routers"
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
	if err := oauth2.RegisterPlatform("gitee", path); err != nil {
		beego.Info(err)
		return
	}

	beego.Run()
}
