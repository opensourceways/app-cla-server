package dbmodels

type IndividualSigningBasicInfo struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Date    string `json:"date"`
	Enabled bool   `json:"enabled"`
}

type IndividualSigningInfo struct {
	IndividualSigningBasicInfo

	CLALanguage string          `json:"cla_language"`
	Info        TypeSigningInfo `json:"info"`
}

type IndividualSigningListOption struct {
	Platform         string `json:"platform"`
	OrgID            string `json:"org_id"`
	RepoID           string `json:"repo_id"`
	CLALanguage      string `json:"cla_language"`
	CorporationEmail string `json:"corporation_email"`
}
