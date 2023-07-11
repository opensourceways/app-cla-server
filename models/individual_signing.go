package models

type IndividualSigning struct {
	Name             string          `json:"name"`
	Email            string          `json:"email"`
	CLAId            string          `json:"cla_id"`
	CLALanguage      string          `json:"cla_language"`
	VerificationCode string          `json:"verification_code"`
	Info             TypeSigningInfo `json:"info"`
}

type IndividualSigningBasicInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Date    string `json:"date"`
	Enabled bool   `json:"enabled"`
}

type IndividualSigningInfo struct {
	IndividualSigningBasicInfo

	CLAId       string          `json:"cla_id"`
	CLALanguage string          `json:"cla_language"`
	Info        TypeSigningInfo `json:"info"`
}
