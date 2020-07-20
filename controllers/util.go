package controllers

import (
	"github.com/astaxie/beego"
)

const (
	headerToken        = "TOKEN"
	headerRefreshToken = "REFRESH_TOKEN"
	headerUser         = "USER"
)

func sendResponse(c *beego.Controller, statusCode int, reason error) {
	c.Ctx.ResponseWriter.WriteHeader(statusCode)

	if reason != nil {
		c.Data["json"] = reason.Error()
	}

	c.ServeJSON()
}

func getHeader(c *beego.Controller, h string) string {
	return c.Ctx.Input.Header(h)
}
