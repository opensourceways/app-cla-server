package models

type IndividualSigning struct {
	CLAOrgID string                 `json:"cla_org_id" required:"true"`
	Email    string                 `json:"email" required:"true"`
	Info     map[string]interface{} `json:"info,omitempty"`
}

func (this *IndividualSigning) Create() error {
	return db.SignAsIndividual(*this)
}
