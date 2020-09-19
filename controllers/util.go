package controllers

import (
	"fmt"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
)

const (
	headerToken          = "Token"
	apiCodePlatformToken = "code_platform_token"
	apiAccessController  = "access_controller"
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

func newAccessToken(user, permission string) (string, error) {
	ac := &accessController{
		User:       user,
		Permission: permission,
		secret:     conf.AppConfig.APITokenKey,
	}

	return ac.NewToken(conf.AppConfig.APITokenExpiry)
}

func newAccessTokenAuthorizedByCodePlatform(user, permission, platformToken string) (string, error) {
	ac := &codePlatformAuth{
		accessController: accessController{
			User:       user,
			Permission: permission,
			secret:     conf.AppConfig.APITokenKey,
		},
		PlatformToken: platformToken,
	}

	return ac.NewToken(conf.AppConfig.APITokenExpiry)
}

func checkApiAccessToken(c *beego.Controller, permission []string, ac accessControllerInterface) (int, error) {
	token := getHeader(c, headerToken)
	if token == "" {
		return 401, fmt.Errorf("no token passed")
	}

	if err := ac.ParseToken(token, conf.AppConfig.APITokenKey); err != nil {
		return 401, err
	}

	if err := ac.Verify(permission); err != nil {
		return 403, err
	}
	return 0, nil
}

func apiPrepare(c *beego.Controller, permission []string, ac accessControllerInterface) {
	if ac == nil {
		ac = &accessController{}
	}

	if code, err := checkApiAccessToken(c, permission, ac); err != nil {
		sendResponse(c, code, err, nil)
		c.StopRun()
	}

	c.Data[apiAccessController] = ac
}

func getAccessController(c *beego.Controller) (accessControllerInterface, error) {
	ac, ok := c.Data[apiAccessController]
	if !ok {
		return nil, fmt.Errorf("no access controller")
	}

	if v, ok := ac.(accessControllerInterface); ok {
		return v, nil
	}

	return nil, fmt.Errorf("can't convert to access controller instance")
}

func getApiAccessUser(c *beego.Controller) (string, error) {
	ac, err := getAccessController(c)
	if err != nil {
		return "", err
	}
	return ac.GetUser(), nil
}

func refreshAccessToken(c *beego.Controller) (string, error) {
	ac, err := getAccessController(c)
	if err != nil {
		return "", err
	}
	return ac.NewToken(conf.AppConfig.APITokenExpiry)
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

func checkAPIStringParameter(c *beego.Controller, params []string) error {
	for _, p := range params {
		if c.GetString(p) == "" {
			return fmt.Errorf("missing parameter of %s", p)
		}
	}
	return nil
}

func checkAndVerifyAPIStringParameter(c *beego.Controller, params map[string]string) error {
	for p, v := range params {
		v1 := c.GetString(p)

		if v1 == "" {
			return fmt.Errorf("missing parameter of %s", p)
		}
		if v != "" && v1 != v {
			return fmt.Errorf("invalid parameter of %s", p)
		}
	}
	return nil
}
