package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
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

type failedResult struct {
	reason     error
	errCode    string
	statusCode int
}

func newFailedResult(statusCode int, errCode string, err error) *failedResult {
	return &failedResult{
		statusCode: statusCode,
		errCode:    errCode,
		reason:     err,
	}
}

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

func apiPrepare(c *beego.Controller, permission []string) {
	var acp interface{}

	switch permission[0] {
	case PermissionOwnerOfOrg:
		acp = &acForCodePlatformPayload{}
	case PermissionIndividualSigner:
		acp = &acForCodePlatformPayload{}
	case PermissionCorporAdmin:
		acp = &acForCorpManagerPayload{}
	case PermissionEmployeeManager:
		acp = &acForCorpManagerPayload{}
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

func getACOfCorpManager(c *beego.Controller) (*acForCorpManagerPayload, string, error) {
	ac, err := getAccessController(c)
	if err != nil {
		return nil, util.ErrInvalidParameter, err
	}

	pl, ok := ac.Payload.(*acForCorpManagerPayload)
	if !ok {
		return nil, util.ErrSystemError, fmt.Errorf("invalid token payload")
	}

	return pl, "", nil
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

func convertDBError(err error) (int, string) {
	e, ok := dbmodels.IsDBError(err)
	if !ok {
		return 500, util.ErrSystemError
	}

	return 400, e.ErrCode
}

func isNoClaBindingDoc(err error) bool {
	_, c := convertDBError(err)
	return c == util.ErrNoCLABindingDoc
}

func getRequestMethod(c *beego.Controller) string {
	return c.Ctx.Request.Method
}

func sendEmailToIndividual(to, from, subject string, builder email.IEmailMessageBulder) {
	sendEmail([]string{to}, from, subject, builder)
}

func sendEmail(to []string, from, subject string, builder email.IEmailMessageBulder) {
	msg, err := builder.GenEmailMsg()
	if err != nil {
		beego.Error(err)
		return
	}

	msg.To = to
	msg.Subject = subject

	worker.GetEmailWorker().SendSimpleMessage(from, msg)
}

func notifyCorpManagerWhenAdding(orgCLA *models.OrgCLA, info []dbmodels.CorporationManagerCreateOption) {
	admin := (info[0].Role == dbmodels.RoleAdmin)
	subject := fmt.Sprintf("Account on project of \"%s\"", orgCLA.OrgAlias)

	for _, item := range info {
		d := email.AddingCorpManager{
			Admin:            admin,
			ID:               item.ID,
			User:             item.Name,
			Email:            item.Email,
			Password:         item.Password,
			Org:              orgCLA.OrgAlias,
			ProjectURL:       projectURL(orgCLA),
			URLOfCLAPlatform: conf.AppConfig.CLAPlatformURL,
		}

		sendEmailToIndividual(item.Email, orgCLA.OrgEmail, subject, d)
	}
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

func getSingingInfo(info dbmodels.TypeSigningInfo, fields []dbmodels.Field) dbmodels.TypeSigningInfo {
	if len(info) == 0 {
		return info
	}

	r := dbmodels.TypeSigningInfo{}
	for _, item := range fields {
		if v, ok := info[item.ID]; ok {
			r[item.ID] = v
		}
	}
	return r
}

func projectURL(orgCLA *models.OrgCLA) string {
	return util.ProjectURL(orgCLA.Platform, orgCLA.RepoID, orgCLA.RepoID)
}

func rspOnAuthFailed(c *beego.Controller, webRedirectDir, errCode string, reason error) {
	setCookies(c, map[string]string{"error_code": errCode, "error_msg": reason.Error()})

	http.Redirect(
		c.Ctx.ResponseWriter, c.Ctx.Request, webRedirectDir, http.StatusFound,
	)
}

func setCookies(c *beego.Controller, value map[string]string) {
	for k, v := range value {
		c.Ctx.SetCookie(k, v, "3600", "/")
	}
}

func downloadFile(c *beego.Controller, path string) {
	c.Ctx.Output.Download(path)
}

func parseOrgAndRepo(s string) (string, string) {
	v := strings.Split(s, ":")
	if len(v) == 2 {
		return v[0], v[1]
	}
	return s, ""
}
