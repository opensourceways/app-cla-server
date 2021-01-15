package dbmodels

type IndividualSigningBasicInfo struct {
	ID      string `json:"id"`
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
