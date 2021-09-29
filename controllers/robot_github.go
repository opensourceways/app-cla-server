package controllers

import (
	"fmt"

	"github.com/astaxie/beego/logs"

	"github.com/opensourceways/app-cla-server/robot/github"
)

type RobotGithubController struct {
	baseController
}

func (this *RobotGithubController) Prepare() {
	this.stopRunIfRobotSerivceIsUnabled()
}

// @Title Post
// @Description retrieving the password by sending an email to the user
// @Param 	link_id		path 	string				true		"link id"
// @Param	body		body 	models.PasswordRetrievalKey	true		"body for retrieving password"
// @Success 201 {string}
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 error_parsing_api_body:     parse payload of request failed
// @Failure 402 no_link:                    the link id is not exists
// @Failure 403 missing_email:              missing email in payload
// @Failure 500 system_error:               system error
// @router / [post]
func (this *RobotGithubController) Post() {
	action := "handle webhook from github"
	payload := this.Ctx.Input.RequestBody

	eventType, eventGUID, code, err := github.ValidateWebhook(payload, this.apiReqHeader)
	if err != nil {
		this.sendFailedResponse(code, "", err, action)
	} else {
		this.sendSuccessResp("Event received. Have a nice day.")
	}

	err = github.Handle(eventType, payload)
	logs.Info(fmt.Sprintf("event type:%s, event id:%s, err:%v", eventType, eventGUID, err))
}
