package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

const (
	headerToken         = "Token"
	apiAccessController = "access_controller"
)

func sendResponse(c *beego.Controller, statusCode, errCode int, reason error, body interface{}, doWhat string) {
	if token, err := refreshAccessToken(c); err == nil {
		// this code must run before `c.Ctx.ResponseWriter.WriteHeader`
		// otherwise the header can't be set successfully.
		// The reason is relevant to the variable of 'Response.Started' at
		// beego/context/context.go
		c.Ctx.Output.Header(headerToken, token)
	}

	f := func(data interface{}) {
		c.Data["json"] = struct {
			Data interface{} `json:"data"`
		}{
			Data: data,
		}
	}

	if reason != nil {
		reason = fmt.Errorf("Failed to %s, err: %s", doWhat, reason.Error())

		if statusCode >= 500 {
			beego.Error(reason.Error())
			reason = fmt.Errorf("System error")
		}

		d := struct {
			ErrCode string `json:"error_code"`
			ErrMsg  string `json:"error_message"`
		}{
			ErrCode: fmt.Sprintf("cla.%04d", errCode),
			ErrMsg:  reason.Error(),
		}

		f(d)

		// if success, don't set status code, otherwise the header set in c.ServeJSON
		// will not work. The reason maybe the same as above.
		c.Ctx.ResponseWriter.WriteHeader(statusCode)
	} else {
		if body != nil {
			f(body)
		}
	}

	c.ServeJSON()
}

func sendResponse1(c *beego.Controller, statusCode int, reason error, body interface{}) {
	if token, err := refreshAccessToken(c); err == nil {
		// this code must run before `c.Ctx.ResponseWriter.WriteHeader`
		// otherwise the header can't be set successfully.
		// The reason is relevant to the variable of 'Response.Started' at
		// beego/context/context.go
		c.Ctx.Output.Header(headerToken, token)
	}

	f := func(data interface{}) {
		c.Data["json"] = struct {
			Data interface{} `json:"data"`
		}{
			Data: data,
		}
	}

	if reason != nil {
		es := reason.Error()
		if statusCode >= 500 {
			beego.Error(es)
			es = "internal error"
		}
		f(es)

		// if success, don't set status code, otherwise the header set in c.ServeJSON
		// will not work. The reason maybe the same as above.
		c.Ctx.ResponseWriter.WriteHeader(statusCode)
	} else {
		if body != nil {
			f(body)
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

func checkApiAccessToken(c *beego.Controller, permission []string, ac accessControllerInterface) (int, int, error) {
	token := getHeader(c, headerToken)
	if token == "" {
		return 401, util.ErrMissingToken, fmt.Errorf("no token passed")
	}

	if err := ac.ParseToken(token, conf.AppConfig.APITokenKey); err != nil {
		return 401, util.ErrUnknownToken, err
	}

	if err := ac.Verify(permission); err != nil {
		return 403, util.ErrInvalidToken, err
	}
	return 0, 0, nil
}

func apiPrepare(c *beego.Controller, permission []string, ac accessControllerInterface) {
	if ac == nil {
		ac = &accessController{}
	}

	if statusCode, errCode, err := checkApiAccessToken(c, permission, ac); err != nil {
		sendResponse(c, statusCode, errCode, err, nil, "")
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

func fetchInputPayload(c *beego.Controller, info interface{}) error {
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, info); err != nil {
		return fmt.Errorf("invalid parameter: %s", err.Error())
	}
	return nil
}

func fetchStringParameter(c *beego.Controller, param string) (string, error) {
	v := c.GetString(param)
	if v == "" {
		return "", fmt.Errorf("missing parameter of %s", param)
	}
	return v, nil
}

func getEmailConfig(claOrgID string) (*models.CLAOrg, *models.OrgEmail, error) {
	claOrg := &models.CLAOrg{ID: claOrgID}
	if err := claOrg.Get(); err != nil {
		return nil, nil, err
	}

	emailInfo := &models.OrgEmail{Email: claOrg.OrgEmail}
	if err := emailInfo.Get(); err != nil {
		return nil, nil, err
	}

	return claOrg, emailInfo, nil
}

func isSameCorp(c *beego.Controller, email string) (int, int, error) {
	corpEmail, err := getApiAccessUser(c)
	if err != nil {
		return 500, 0, err
	}

	if util.EmailSuffix(corpEmail) != util.EmailSuffix(email) {
		return 400, util.ErrInvalidParameter, fmt.Errorf("can't operate on the different corporation")
	}

	return 0, 0, nil
}

func convertDBError(err error) (int, int) {
	e, ok := dbmodels.IsDBError(err)
	if !ok {
		return 500, 0
	}

	return 400, e.ErrCode
}
