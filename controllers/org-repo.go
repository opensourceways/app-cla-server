package controllers

type OrgRepoController struct {
	baseController
}

func (org *OrgRepoController) Prepare() {
	org.apiPrepare(PermissionOwnerOfOrg)
}

// @Title List
// @Description get all orgs
// @Success 200 {string} list
// @Failure 400 missing_token:              token is missing
// @Failure 401 unknown_token:              token is unknown
// @Failure 402 expired_token:              token is expired
// @Failure 403 unauthorized_token:         the permission of token is unmatched
// @Failure 500 system_error:               system error
// @router / [get]
func (org *OrgRepoController) List() {
	action := "list org"

	pl, fr := org.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		org.sendFailedResultAsResp(fr, action)
		return
	}

	pl.refreshOrg()

	r := make([]string, 0, len(pl.Orgs))
	for k := range pl.Orgs {
		r = append(r, k)
	}

	org.sendSuccessResp(r)
}

// @Title Check
// @Description check whether the repo exists
// @Param	:org	path 	string				true		"org"
// @Param	:repo	path 	string				true		"repo"
// @Success 200 {string} map
// @Failure 400 missing_url_path_parameter: missing url path parameter
// @Failure 401 missing_token:              token is missing
// @Failure 402 unknown_token:              token is unknown
// @Failure 403 expired_token:              token is expired
// @Failure 404 unauthorized_token:         the permission of token is unmatched
// @Failure 500 system_error:               system error
// @router /:org/:repo [get]
func (org *OrgRepoController) Check() {
	action := "check repo"

	pl, fr := org.tokenPayloadBasedOnCodePlatform()
	if fr != nil {
		org.sendFailedResultAsResp(fr, action)
		return
	}

	repo := org.GetString(":repo")
	b, err := pl.hasRepo(org.GetString(":org"), repo)
	if err != nil {
		org.sendFailedResultAsResp(err, action)
		return
	}

	org.sendSuccessResp(b)
}
