package controllers

import (
	"strings"

	"github.com/astaxie/beego"
)

const (
	headerAccessToken  = "Access-Token"
	headerRefreshToken = "Refresh-Token"
	headerUser         = "User"
)

func sendResponse(c *beego.Controller, statusCode int, reason error, body interface{}) {
	c.Ctx.ResponseWriter.WriteHeader(statusCode)

	if reason != nil {
		c.Data["json"] = reason.Error()
	} else {
		if body != nil {
			c.Data["json"] = body
		}
	}

	c.ServeJSON()
}

func getHeader(c *beego.Controller, h string) string {
	return c.Ctx.Input.Header(h)
}

type requestHeader struct {
	accessToken  string
	refreshToken string
	platform     string
	user         string
}

func parseHeader(c *beego.Controller) requestHeader {
	h := requestHeader{
		accessToken:  c.Ctx.Input.Header(headerAccessToken),
		refreshToken: c.Ctx.Input.Header(headerRefreshToken),
		user:         c.Ctx.Input.Header(headerUser),
	}

	v := strings.Split(h.user, "/")
	if len(v) == 2 {
		h.platform = v[0]
	}

	return h
}
