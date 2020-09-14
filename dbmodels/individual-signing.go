package dbmodels

type IndividualSigningInfo struct {
	Email   string          `json:"email" required:"true"`
	Name    string          `json:"name" required:"true"`
	Enabled bool            `json:"enabled"`
	Info    TypeSigningInfo `json:"info,omitempty"`
}
