package controllers

import "github.com/opensourceways/app-cla-server/models"

type MigrationController struct {
	baseController
}

func (ctl *MigrationController) Prepare() {
	ctl.apiPrepare(PermissionOwnerOfOrg)
}

// @Title MigrateCommunityData
// @Description migrate CLA data from old community to new community
// @Tags Migration
// @Accept json
// @Param body body models.CommunityMigrationOpt true "migration options"
// @Success 200 {object} controllers.respData
// @router / [post]
func (ctl *MigrationController) MigrateCommunityData() {
	action := "migrate community CLA data"

	pl, fr := ctl.tokenPayloadBasedOnCorpManager()
	if fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	input := &models.CommunityMigrationOpt{}
	if fr := ctl.fetchInputPayload(input); fr != nil {
		ctl.sendFailedResultAsResp(fr, action)
		return
	}

	if err := models.MigrateCommunityData(pl.UserId, input); err != nil {
		ctl.sendModelErrorAsResp(err, action)
	} else {
		ctl.sendSuccessResp(action, "successfully")
	}
}
