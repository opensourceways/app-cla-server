package controllers

type OrganizationController struct {
	baseController
}

func (ctl *OrganizationController) Prepare() {
	ctl.apiPrepare(PermissionOwnerOfOrg)
}

// @Title ListOrganizations
// @Description list all organizations
// @Success 200 {} map
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 500 system_error:               system error
// @router / [get]
func (ctl *OrganizationController) List() {
	if pl, fr := ctl.tokenPayloadBasedOnCodePlatform(); fr != nil {
		ctl.sendFailedResultAsResp(fr, "list organizations")
	} else {
		ctl.sendSuccessResp(pl.Orgs)
	}
}
