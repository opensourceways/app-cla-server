package dbmodels

type IndividualSigningInfo struct {
	Email   string          `json:"email"`
	Name    string          `json:"name"`
	Date    string          `json:"date"`
	Enabled bool            `json:"enabled"`
	Info    TypeSigningInfo `json:"info"`
}
