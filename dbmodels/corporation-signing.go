package dbmodels

type TypeSigningInfo map[string]string

type CorporationSigningBasicInfo struct {
	CLALanguage     string `json:"cla_language"`
	AdminEmail      string `json:"admin_email"`
	AdminName       string `json:"admin_name"`
	CorporationName string `json:"corporation_name"`
	Date            string `json:"date"`
}

type CorporationSigningSummary struct {
	CorporationSigningBasicInfo

	AdminAdded bool `json:"admin_added"`
}

type CorporationSigningOption struct {
	CorporationSigningBasicInfo

	Info TypeSigningInfo `json:"info"`
}
