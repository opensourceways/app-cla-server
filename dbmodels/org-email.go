package dbmodels

type OrgEmailCreateInfo struct {
	Email string `json:"email" required:"true"`
	Token []byte `json:"token" required:"true"`
}
