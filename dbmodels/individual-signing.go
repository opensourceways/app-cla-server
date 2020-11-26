package dbmodels

type IndividualSigningBasicInfo struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Date    string `json:"date"`
	Enabled bool   `json:"enabled"`
}

type IndividualSigningInfo struct {
	IndividualSigningBasicInfo

	Info TypeSigningInfo `json:"info"`
}

type IndividualSigningListOption struct {
	CLALanguage      string `json:"cla_language"`
	CorporationEmail string `json:"corporation_email"`
}
