package main

import (
	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/dbmodels"
	"github.com/zengchen1024/cla-server/email"
	"github.com/zengchen1024/cla-server/models"
	"github.com/zengchen1024/cla-server/mongodb"
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
	if err = email.NewGmailClient(path); err != nil {
		return
	}

	beego.Run()
}
