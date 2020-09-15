package controllers

import (
	"fmt"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
)

const (
	headerToken          = "Token"
	apiAccessUser        = "access_user"
	apiCodePlatformToken = "code_platform_token"
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

func checkApiAccessToken(c *beego.Controller, permission []string, ac accessControllerInterface) error {
	token := getHeader(c, headerToken)
	if token == "" {
		return fmt.Errorf("no token passed")
	}

	return ac.CheckToken(token, conf.AppConfig.APITokenKey, permission)
}

func apiPrepare(c *beego.Controller, permission []string, ac accessControllerInterface) {
	if ac == nil {
		ac = &accessController{}
	}

	if err := checkApiAccessToken(c, permission, ac); err != nil {
		sendResponse(c, 400, err, nil)
		c.StopRun()
	}

	c.Data[apiAccessUser] = ac.GetUser()
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
	case dbmodels.RoleAdmin:
		return PermissionCorporAdmin
	case dbmodels.RoleManager:
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

func getRouterPattern(c *beego.Controller) string {
	v, ok := c.Data["RouterPattern"]
	if ok {
		return v.(string)
	}
	return ""
}
