package models

type EmployeeSigning struct {
	CLAOrgID string                 `json:"cla_org_id" required:"true"`
	Email    string                 `json:"email" required:"true"`
	Info     map[string]interface{} `json:"info,omitempty"`
}

func (this *EmployeeSigning) Create() error {
	return db.SignAsEmployee(*this)
}
