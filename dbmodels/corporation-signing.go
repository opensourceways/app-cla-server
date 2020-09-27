package dbmodels

type TypeSigningInfo map[string]string

type CorporationSigningBasicInfo struct {
	AdminEmail      string `json:"admin_email"`
	AdminName       string `json:"admin_name"`
	CorporationName string `json:"corporation_name"`
	Date            string `json:"date"`
}

type CorporationSigningDetail struct {
	CorporationSigningBasicInfo

	PDFUploaded bool `json:"pdf_uploaded"`
	AdminAdded  bool `json:"admin_added"`
}

type CorporationSigningInfo struct {
	CorporationSigningBasicInfo

	Info TypeSigningInfo `json:"info"`
}

type CorporationSigningListOption struct {
	Platform    string `json:"platform"`
	OrgID       string `json:"org_id"`
	RepoID      string `json:"repo_id"`
	CLALanguage string `json:"cla_language"`
}

type CorporationSigningUpdateInfo struct {
	Enabled *bool `json:"enabled,omitempty"`
}
