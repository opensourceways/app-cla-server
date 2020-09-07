package dbmodels

type IndividualSigningInfo struct {
	Email string          `json:"email" required:"true"`
	Info  TypeSigningInfo `json:"info,omitempty"`
}
