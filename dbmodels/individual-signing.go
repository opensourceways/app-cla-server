package dbmodels

type IndividualSigningInfo struct {
	Email string                 `json:"email" required:"true"`
	Info  map[string]interface{} `json:"info,omitempty"`
}
