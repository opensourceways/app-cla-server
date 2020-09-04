package controllers

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/models"
)

const (
	headerAccessToken  = "Access-Token"
	headerRefreshToken = "Refresh-Token"
	headerUser         = "User"
	headerToken        = "Token"
	apiAccessUser      = "access_user"
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

func createApiAccessToken(user, permission string) (string, error) {
	expiry, err := beego.AppConfig.Int64("api_token_expiry")
	if err != nil {
		return "", fmt.Errorf("Failed to create access token: parsing token expiry was failed")
	}
	ac := &accessControler{
		User:       user,
		Permission: permission,
	}
	return ac.CreateToken(
		expiry,
		beego.AppConfig.String("api_token_key"),
	)
}

func checkApiAccessToken(c *beego.Controller, permission []string) (string, error) {
	token := getHeader(c, headerToken)
	if token == "" {
		return "", fmt.Errorf("no token passed")
	}

	ac := &accessControler{}

	err := ac.CheckToken(token, beego.AppConfig.String("api_token_key"), permission)
	if err != nil {
		return "", err
	}

	return ac.User, nil
}

func apiPrepare(c *beego.Controller, permission []string) {
	user, err := checkApiAccessToken(c, permission)
	if err != nil {
		sendResponse(c, 400, err, nil)
		c.StopRun()
	}

	c.Data[apiAccessUser] = user
}

func getApiAccessUser(c *beego.Controller) (string, error) {
	user, ok := c.Data[apiAccessUser].(string)
	if !ok {
		return "", fmt.Errorf("no user")
	}
	return user, nil
}

func corporRoleToPermission(role string) string {
	switch role {
	case models.RoleAdmin:
		return PermissionCorporAdmin
	case models.RoleManager:
		return PermissionEmployeeManager
	}
	return ""
}

func actionToPermission(action string) string {
	switch action {
	case "login":
		return PermissionOwnerOfOrg
	case "sign":
		return PermissionIndividualSigner
	}
	return ""
}
