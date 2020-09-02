package dbmodels

type VerificationCode struct {
	Email   string `json:"email" required:"true"`
	Code    string `json:"code" required:"true"`
	Purpose string `json:"purpose" required:"true"`
	Expiry  int64  `json:"expiry" required:"true"`
}
