package controllers

type HeartbeatController struct {
	baseController
}

// @Title Heartbeat
// @Description  heartbeat
// @Tags Heartbeat
// @Accept json
// @Success 200 {object} controllers.respData
// @router / [get]
func (ctl *HeartbeatController) Heartbeat() {
	ctl.sendSuccessResp("heartbeat", "good")
}
