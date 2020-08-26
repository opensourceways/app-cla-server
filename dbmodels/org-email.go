package dbmodels

type OrgEmailCreateInfo struct {
	Email    string `json:"email" required:"true"`
	Platform string `json:"platform" required:"true"`
	Token    []byte `json:"token" required:"true"`
}
