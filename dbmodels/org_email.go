package dbmodels

type OrgEmailCreateInfo struct {
	Email    string
	Platform string
	Token    []byte
	AuthCode string
}
