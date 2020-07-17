package controllers

import (
	"github.com/astaxie/beego"
)

func sendResponse(c *beego.Controller, statusCode int, reason error) {
	c.Ctx.ResponseWriter.WriteHeader(statusCode)

	if reason != nil {
		c.Data["json"] = reason.Error()
	}

	c.ServeJSON()
}
