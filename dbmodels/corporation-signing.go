package dbmodels

type TypeSigningInfo map[string]string

type SigningIndex struct {
	LinkId    string
	SigningId string
}

type CorporationSigningBasicInfo struct {
	ID              string `json:"id"`
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

type CorpSigningCreateOpt struct {
	CorporationSigningBasicInfo

	Info TypeSigningInfo `json:"info"`
}

type CorpSigningListOpt struct {
	Lang         string
	EmailDomain  string
	IncludeAdmin bool
}
