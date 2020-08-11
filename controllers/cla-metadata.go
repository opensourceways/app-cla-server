package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"

	"github.com/zengchen1024/cla-server/models"
)

type CLAMetadataController struct {
	beego.Controller
}

// @Title CreateCLAMetadata
// @Description create cla metadata
// @Param	body		body 	models.CLAMetadata	true		"body for cla metadata"
// @Success 201 {int} models.CLAMetadata
// @Failure 403 body is empty
// @router / [post]
func (this *CLAMetadataController) Post() {
	var statusCode = 201
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	var data models.CLAMetadata
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &data); err != nil {
		reason = err
		statusCode = 400
		return
	}

	submitter := getHeader(&this.Controller, headerUser)
	data.Submitter = submitter

	if err := (&data).Create(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = data
}

// @Title Delete CLAMetadata
// @Description delete cla metadata
// @Param	uid		path 	string	true		"cla metadata id"
// @Success 204 {string} delete success!
// @Failure 403 uid is empty
// @router /:uid [delete]
func (this *CLAMetadataController) Delete() {
	var statusCode = 204
	var reason error
	var body string

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	uid := this.GetString(":uid")
	if uid == "" {
		reason = fmt.Errorf("missing cla metadata id")
		statusCode = 400
		return
	}

	data := models.CLAMetadata{ID: uid}

	if err := (&data).Delete(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = "delete cla metadata successfully"
}

// @Title Get
// @Description get cla metadata by uid
// @Param	uid		path 	string	true		"The key for cla metadata"
// @Success 200 {object} models.CLAMetadata
// @Failure 403 :uid is empty
// @router /:uid [get]
func (this *CLAMetadataController) Get() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	uid := this.GetString(":uid")
	if uid == "" {
		reason = fmt.Errorf("missing cla metadata id")
		statusCode = 400
		return
	}

	data := models.CLAMetadata{ID: uid}

	if err := (&data).Get(); err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = data
}

// @Title GetAllCLAMetadata
// @Description get all cla metadatas
// @Success 200 {object} models.CLAMetadata
// @router / [get]
func (this *CLAMetadataController) GetAll() {
	var statusCode = 200
	var reason error
	var body interface{}

	defer func() {
		sendResponse(&this.Controller, statusCode, reason, body)
	}()

	datas := models.CLAMetadatas{BelongTo: []string{getHeader(&this.Controller, headerUser)}}

	r, err := datas.Get()
	if err != nil {
		reason = err
		statusCode = 500
		return
	}

	body = r
}
