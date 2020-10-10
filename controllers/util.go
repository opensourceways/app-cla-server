package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/conf"
	"github.com/opensourceways/app-cla-server/dbmodels"
	"github.com/opensourceways/app-cla-server/email"
	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
	"github.com/opensourceways/app-cla-server/worker"
)

const (
	headerToken         = "Token"
	apiAccessController = "access_controller"
)

func buildStatusAndErrCode(statusCode int, errCode string, reason error) (int, string) {
	if errCode == "" {
		sc, ec := convertDBError(reason)
		if statusCode == 0 {
			return sc, ec
		}
		return statusCode, ec
	}

	if statusCode == 0 {
		return 500, errCode
	}

	return statusCode, errCode
}

func sendResponse(c *beego.Controller, statusCode int, errCode string, reason error, body interface{}, doWhat string) {
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
		statusCode, errCode := buildStatusAndErrCode(statusCode, errCode, reason)

		if statusCode >= 500 {
			beego.Error(fmt.Sprintf("Failed to %s, errCode: %s, err: %s", doWhat, errCode, reason.Error()))

			reason = fmt.Errorf("System error")
			errCode = util.ErrSystemError
		}

		d := struct {
			ErrCode string `json:"error_code"`
			ErrMsg  string `json:"error_message"`
		}{
			ErrCode: fmt.Sprintf("cla.%s", errCode),
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

func checkApiAccessToken(c *beego.Controller, permission []string, ac accessControllerInterface) (int, string, error) {
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
	return 0, "", nil
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

func isSameCorp(email1, email2 string) bool {
	return util.EmailSuffix(email1) == util.EmailSuffix(email2)
}

func checkSameCorp(c *beego.Controller, email string) (int, string, error) {
	_, corpEmail, err := parseCorpManagerUser(c)
	if err != nil {
		return 401, util.ErrUnknownToken, err
	}

	if !isSameCorp(corpEmail, email) {
		return 400, util.ErrNotSameCorp, fmt.Errorf("not same corp")
	}

	return 0, "", nil
}

func convertDBError(err error) (int, string) {
	e, ok := dbmodels.IsDBError(err)
	if !ok {
		return 500, ""
	}

	return 400, e.ErrCode
}

func corpManagerUser(claOrgID, email string) string {
	return fmt.Sprintf("%s/%s", claOrgID, email)
}

func parseCorpManagerUser(c *beego.Controller) (string, string, error) {
	user, err := getApiAccessUser(c)
	if err != nil {
		return "", "", err
	}

	v := strings.Split(user, "/")
	if len(v) != 2 {
		return "", "", fmt.Errorf("can't parse corp manager user")
	}

	return v[0], v[1], nil
}

func isNoClaBindingDoc(err error) bool {
	_, c := convertDBError(err)
	return c == util.ErrNoCLABindingDoc
}

func getRequestMethod(c *beego.Controller) string {
	return c.Ctx.Request.Method
}

func notifyCorpManagerWhenAdding(orgEmail, subject string, info []dbmodels.CorporationManagerCreateOption) {
	for _, item := range info {
		d := email.AddingCorpManager{
			Admin:    (item.Role == dbmodels.RoleAdmin),
			Password: item.Password,
		}
		msg, err := d.GenEmailMsg()
		if err != nil {
			beego.Error(err)
			continue
		}
		msg.To = []string{item.Email}
		msg.Subject = subject

		worker.GetEmailWorker().SendSimpleMessage(orgEmail, msg)
	}
}

func notifyCorpManagerWhenRemoving(orgEmail string, info []string) {
	for _, item := range info {
		d := email.RemovingCorpManager{}
		msg, err := d.GenEmailMsg()
		if err != nil {
			beego.Error(err)
			continue
		}

		msg.To = []string{item}
		msg.Subject = "Removing Corp Manager"

		worker.GetEmailWorker().SendSimpleMessage(orgEmail, msg)
	}
}

func sendVerificationCodeEmail(code, orgEmail, adminEmail string) {
	d := email.CorpSigningVerificationCode{Code: code}
	msg, err := d.GenEmailMsg()
	if err != nil {
		beego.Error(err)
		return
	}

	msg.To = []string{adminEmail}
	msg.Subject = "verification code"

	worker.GetEmailWorker().SendSimpleMessage(orgEmail, msg)
}
