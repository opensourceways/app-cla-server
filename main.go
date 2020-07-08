package main

import (
	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla/models"
	"github.com/zengchen1024/cla/mongodb"
	_ "github.com/zengchen1024/cla/routers"
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

	beego.Run()
}
