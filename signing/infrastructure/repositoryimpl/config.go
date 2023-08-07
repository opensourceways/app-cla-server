package repositoryimpl

type Config struct {
	Collections Collections `json:"collections" required:"true"`
}

type Collections struct {
	Org               string `json:"org"                required:"true"`
	CLA               string `json:"cla"                required:"true"`
	Link              string `json:"link"               required:"true"`
	User              string `json:"user"               required:"true"`
	CorpSigning       string `json:"corp_signing"       required:"true"`
	EmailCredential   string `json:"email_credential"   required:"true"`
	VerificationCode  string `json:"verification_code"  required:"true"`
	IndividualSigning string `json:"individual_signing" required:"true"`
}
