package dbmodels

type EmployeeSigningInfo struct {
	Email   string `json:"email" required:"true"`
	Name    string `json:"name" required:"true"`
	Enabled bool   `json:"enabled"`

	Info map[string]interface{} `json:"info,omitempty"`
}
