package dbmodels

type VerificationCode struct {
	Email   string
	Code    string
	Purpose string
	Expiry  int64
}
