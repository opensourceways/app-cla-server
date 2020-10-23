package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/astaxie/beego"

	"github.com/opensourceways/app-cla-server/code-platform-auth/platforms"
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

func checkApiAccessToken(c *beego.Controller, permission []string, ac *accessController) (int, string, error) {
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

func apiPrepare(c *beego.Controller, permission []string, acp interface{}) {
	if acp == nil {
		acp = &accessControllerBasicPayload{}
	}

	ac := &accessController{
		Payload: acp,
	}

	if statusCode, errCode, err := checkApiAccessToken(c, permission, ac); err != nil {
		sendResponse(c, statusCode, errCode, err, nil, "")
		c.StopRun()
	}

	c.Data[apiAccessController] = ac
}

func getAccessController(c *beego.Controller) (*accessController, error) {
	ac, ok := c.Data[apiAccessController]
	if !ok {
		return nil, fmt.Errorf("no access controller")
	}

	if v, ok := ac.(*accessController); ok {
		return v, nil
	}

	return nil, fmt.Errorf("can't convert to access controller instance")
}

func refreshAccessToken(c *beego.Controller) (string, error) {
	ac, err := getAccessController(c)
	if err != nil {
		return "", err
	}
	return ac.RefreshToken(conf.AppConfig.APITokenExpiry, conf.AppConfig.APITokenKey)
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

func getEmailConfig(orgCLAID string) (*models.OrgCLA, *models.OrgEmail, error) {
	orgCLA := &models.OrgCLA{ID: orgCLAID}
	if err := orgCLA.Get(); err != nil {
		return nil, nil, err
	}

	emailInfo := &models.OrgEmail{Email: orgCLA.OrgEmail}
	if err := emailInfo.Get(); err != nil {
		return nil, nil, err
	}

	return orgCLA, emailInfo, nil
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
		return 500, util.ErrSystemError
	}

	return 400, e.ErrCode
}

func corpManagerUser(orgCLAID, email string) string {
	return fmt.Sprintf("%s/%s", orgCLAID, email)
}

func parseCorpManagerUser(c *beego.Controller) (string, string, error) {
	ac, err := getAccessController(c)
	if err != nil {
		return "", "", err
	}

	p, ok := ac.Payload.(*accessControllerBasicPayload)
	if !ok {
		return "", "", fmt.Errorf("fetch token Payload failed")
	}

	v := strings.Split(p.User, "/")
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

func isNotIndividualCLA(orgCLA *models.OrgCLA) bool {
	return orgCLA.ApplyTo != dbmodels.ApplyToIndividual
}

func isNotCorpCLA(orgCLA *models.OrgCLA) bool {
	return orgCLA.ApplyTo != dbmodels.ApplyToCorporation
}

func canAccessOrgCLA(c *beego.Controller, orgCLAID string) (*models.OrgCLA, int, string, error) {
	orgCLA := &models.OrgCLA{ID: orgCLAID}
	if err := orgCLA.Get(); err != nil {
		return nil, 400, util.ErrInvalidParameter, err
	}

	ac, ec, err := getACOfCodePlatform(c)
	if err != nil {
		return nil, 400, ec, err
	}

	org := orgCLA.OrgID
	if ac.hasOrg(org) {
		return orgCLA, 0, "", nil
	}

	p, err := platforms.NewPlatform(ac.PlatformToken, "", ac.Platform)
	if err != nil {
		return nil, 400, util.ErrInvalidParameter, err
	}

	b, err := p.IsOrgExist(org)
	if err != nil {
		// TODO token expiry
		return nil, 500, util.ErrSystemError, err
	}

	if !b {
		return nil, 400, util.ErrNotYoursOrg, fmt.Errorf("not the org of owner")
	}

	ac.addOrg(org)

	return orgCLA, 0, "", nil
}

func getACOfCodePlatform(c *beego.Controller) (*acForCodePlatformPayload, string, error) {
	ac, err := getAccessController(c)
	if err != nil {
		return nil, util.ErrInvalidParameter, err
	}

	cpa, ok := ac.Payload.(*acForCodePlatformPayload)
	if !ok {
		return nil, util.ErrSystemError, fmt.Errorf("invalid token payload")
	}

	return cpa, "", nil
}

func trimSingingInfo(info dbmodels.TypeSigningInfo, fields []dbmodels.Field) {
	if len(info) == 0 {
		return
	}

	m := map[string]bool{}
	for _, item := range fields {
		m[item.ID] = true
	}

	for k := range info {
		if _, ok := m[k]; !ok {
			delete(info, k)
		}
	}
}
