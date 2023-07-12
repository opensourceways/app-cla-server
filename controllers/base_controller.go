package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	beego "github.com/beego/beego/v2/server/web"

	"github.com/opensourceways/app-cla-server/models"
	"github.com/opensourceways/app-cla-server/util"
)

type failedApiResult struct {
	reason     error
	errCode    string
	statusCode int
}

func newFailedApiResult(statusCode int, errCode string, err error) *failedApiResult {
	return &failedApiResult{
		statusCode: statusCode,
		errCode:    errCode,
		reason:     err,
	}
}

type baseController struct {
	beego.Controller
}

func (ctl *baseController) fetchInputPayload(info interface{}) *failedApiResult {
	return fetchInputPayloadData(ctl.Ctx.Input.RequestBody, info)
}

func (ctl *baseController) fetchInputPayloadFromFormData(info interface{}) *failedApiResult {
	input := []byte(ctl.Ctx.Request.FormValue("data"))
	return fetchInputPayloadData(input, info)
}

func (ctl *baseController) checkPathParameter() *failedApiResult {
	rp := ctl.routerPattern()
	if rp == "" {
		return nil
	}

	items := strings.Split(rp, "/")
	for _, item := range items {
		if strings.HasPrefix(item, ":") && ctl.GetString(item) == "" {
			return newFailedApiResult(
				400, errMissingURLPathParameter,
				fmt.Errorf("missing path parameter:%s", item))
		}
	}

	return nil
}

func (ctl *baseController) routerPattern() string {
	if v, ok := ctl.Data["RouterPattern"]; ok {
		return v.(string)
	}
	return ""
}

func (ctl *baseController) apiReqHeader(h string) string {
	return ctl.Ctx.Input.Header(h)
}

func (ctl *baseController) apiRequestMethod() string {
	return ctl.Ctx.Request.Method
}

func (ctl *baseController) isPostRequest() bool {
	return ctl.apiRequestMethod() == http.MethodPost
}

func (ctl *baseController) isPutRequest() bool {
	return ctl.apiRequestMethod() == http.MethodPut
}

func (ctl *baseController) isGetRequest() bool {
	return ctl.apiRequestMethod() == http.MethodGet
}

func (ctl *baseController) readInputFile(fileName string, maxSize int, fileType string) ([]byte, *failedApiResult) {
	if v := ctl.Ctx.Request.ContentLength; v <= 0 || v > int64(maxSize) {
		return nil, newFailedApiResult(400, errTooBigPDF, fmt.Errorf("big pdf file"))
	}

	f, _, err := ctl.GetFile(fileName)
	if err != nil {
		return nil, newFailedApiResult(400, errReadingFile, err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, newFailedApiResult(500, errSystemError, err)
	}

	if len(data) > maxSize {
		return nil, newFailedApiResult(400, errTooBigPDF, fmt.Errorf("big pdf file"))
	}

	if !util.CheckContentType(data, fileType) {
		return nil, newFailedApiResult(400, errWrongFileType, fmt.Errorf("unsupported file type"))
	}

	return data, nil
}

func (ctl *baseController) downloadFile(path string) {
	ctl.Ctx.Output.Download(path)
}

func (ctl *baseController) redirect(webRedirectDir string) {
	http.Redirect(
		ctl.Ctx.ResponseWriter, ctl.Ctx.Request, webRedirectDir, http.StatusFound,
	)
}

func (ctl *baseController) logout() {
	if t := ctl.Ctx.GetCookie(accessToken); t != "" {
		models.RemoveAccessToken(t)
	}
	ctl.setToken(models.AccessToken{})
}
