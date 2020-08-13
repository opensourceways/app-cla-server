package models

type CorporationSigning struct {
	CLAOrgID string                 `json:"cla_org_id" required:"true"`
	Email    string                 `json:"email" required:"true"`
	Info     map[string]interface{} `json:"info,omitempty"`
}

func (this *CorporationSigning) Create() error {
	return db.SignAsCorporation(*this)
}
